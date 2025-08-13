"""
Training script for DQN agent on LLM attack optimization.

This script handles the training loop, logging, and checkpointing
for the reinforcement learning agent.
"""

import os
import sys
import argparse
import yaml
import json
from datetime import datetime
from pathlib import Path
import numpy as np
import torch
from typing import Dict, Any, Optional
from tqdm import tqdm
import mlflow
import mlflow.pytorch

# Add parent directory to path
sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))))

from ml.environments.attack_env import make_attack_env
from ml.agents.dqn import DQNAgent, DQNConfig


class RLTrainer:
    """Handles training of RL agent for LLM attacks"""
    
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        
        # Create environment
        env_config = config.get('environment', {})
        self.env = make_attack_env(env_config)
        
        # Create agent
        agent_config = DQNConfig(**config.get('agent', {}))
        self.agent = DQNAgent(
            self.env.observation_space,
            self.env.action_space,
            agent_config
        )
        
        # Training parameters
        self.num_episodes = config.get('num_episodes', 1000)
        self.max_steps_per_episode = config.get('max_steps_per_episode', 100)
        self.save_frequency = config.get('save_frequency', 100)
        self.eval_frequency = config.get('eval_frequency', 50)
        self.eval_episodes = config.get('eval_episodes', 10)
        
        # Logging
        self.experiment_name = config.get('experiment_name', 'rl_training')
        self.checkpoint_dir = Path(config.get('checkpoint_dir', 'checkpoints'))
        self.checkpoint_dir.mkdir(parents=True, exist_ok=True)
        
        # Metrics tracking
        self.episode_rewards = []
        self.episode_lengths = []
        self.success_rates = []
        self.losses = []
        
    def train(self):
        """Main training loop"""
        print(f"Starting RL training on {self.device}")
        print(f"Environment: {self.env}")
        print(f"Agent: DQN with config {self.agent.config}")
        
        # Initialize MLflow
        mlflow.set_experiment(self.experiment_name)
        
        with mlflow.start_run():
            # Log configuration
            mlflow.log_params({
                "num_episodes": self.num_episodes,
                "max_steps": self.max_steps_per_episode,
                "device": str(self.device),
                "agent_type": "DQN",
                "learning_rate": self.agent.config.learning_rate,
                "batch_size": self.agent.config.batch_size,
                "gamma": self.agent.config.gamma,
                "epsilon_start": self.agent.config.epsilon_start,
                "epsilon_end": self.agent.config.epsilon_end
            })
            
            # Training loop
            progress_bar = tqdm(range(self.num_episodes), desc="Training")
            
            for episode in progress_bar:
                # Run training episode
                episode_reward, episode_length, success_rate = self._run_episode(training=True)
                
                self.episode_rewards.append(episode_reward)
                self.episode_lengths.append(episode_length)
                self.success_rates.append(success_rate)
                
                # Update progress bar
                progress_bar.set_postfix({
                    'reward': f"{episode_reward:.2f}",
                    'length': episode_length,
                    'success': f"{success_rate:.2%}",
                    'epsilon': f"{self.agent.epsilon:.3f}"
                })
                
                # Log metrics
                mlflow.log_metrics({
                    "episode_reward": episode_reward,
                    "episode_length": episode_length,
                    "success_rate": success_rate,
                    "epsilon": self.agent.epsilon,
                    "beta": self.agent.beta
                }, step=episode)
                
                # Log training loss
                if self.losses:
                    avg_loss = np.mean(self.losses[-100:])
                    mlflow.log_metric("avg_loss", avg_loss, step=episode)
                
                # Evaluation
                if episode % self.eval_frequency == 0 and episode > 0:
                    eval_reward, eval_success = self._evaluate()
                    mlflow.log_metrics({
                        "eval_reward": eval_reward,
                        "eval_success_rate": eval_success
                    }, step=episode)
                    
                    print(f"\n[Eval] Episode {episode}: "
                          f"Reward={eval_reward:.2f}, Success={eval_success:.2%}")
                
                # Save checkpoint
                if episode % self.save_frequency == 0 and episode > 0:
                    self._save_checkpoint(episode)
                    
            # Final evaluation
            final_reward, final_success = self._evaluate()
            print(f"\nFinal evaluation: Reward={final_reward:.2f}, "
                  f"Success={final_success:.2%}")
            
            # Save final model
            self._save_checkpoint("final")
            mlflow.pytorch.log_model(
                self.agent.q_network,
                "final_model",
                pip_requirements=["torch", "numpy"]
            )
            
            # Log summary statistics
            mlflow.log_metrics({
                "final_eval_reward": final_reward,
                "final_eval_success": final_success,
                "avg_training_reward": np.mean(self.episode_rewards),
                "avg_success_rate": np.mean(self.success_rates)
            })
            
    def _run_episode(self, training: bool = True) -> tuple:
        """
        Run a single episode.
        
        Returns:
            episode_reward: Total reward for the episode
            episode_length: Number of steps taken
            success_rate: Success rate of attacks in episode
        """
        state = self.env.reset()
        episode_reward = 0
        episode_length = 0
        successes = []
        
        for step in range(self.max_steps_per_episode):
            # Select action
            action = self.agent.act(state, explore=training)
            
            # Take action
            next_state, reward, done, info = self.env.step(action)
            
            # Store experience
            if training:
                self.agent.remember(state, action, reward, next_state, done)
                
                # Learn from experience
                if len(self.agent.memory) > self.agent.config.batch_size:
                    loss = self.agent.learn()
                    if loss is not None:
                        self.losses.append(loss)
            
            # Track metrics
            episode_reward += reward
            episode_length += 1
            if 'attack_result' in info:
                successes.append(info['attack_result'].success)
            
            # Update state
            state = next_state
            
            if done:
                break
                
        success_rate = np.mean(successes) if successes else 0.0
        
        return episode_reward, episode_length, success_rate
    
    def _evaluate(self) -> tuple:
        """
        Evaluate agent performance without exploration.
        
        Returns:
            avg_reward: Average reward over evaluation episodes
            avg_success: Average success rate
        """
        eval_rewards = []
        eval_successes = []
        
        for _ in range(self.eval_episodes):
            # Run episode without exploration
            reward, _, success = self._run_episode(training=False)
            eval_rewards.append(reward)
            eval_successes.append(success)
            
        return np.mean(eval_rewards), np.mean(eval_successes)
    
    def _save_checkpoint(self, episode: str):
        """Save training checkpoint"""
        checkpoint_path = self.checkpoint_dir / f"checkpoint_{episode}.pt"
        
        # Save agent
        self.agent.save(str(checkpoint_path))
        
        # Save training state
        state_path = self.checkpoint_dir / f"training_state_{episode}.json"
        with open(state_path, 'w') as f:
            json.dump({
                'episode': episode,
                'episode_rewards': self.episode_rewards[-100:],  # Last 100
                'success_rates': self.success_rates[-100:],
                'timestamp': datetime.now().isoformat()
            }, f, indent=2)
            
        print(f"\nCheckpoint saved: {checkpoint_path}")


def load_config(config_path: str) -> Dict[str, Any]:
    """Load configuration from YAML file"""
    with open(config_path, 'r') as f:
        config = yaml.safe_load(f)
    return config


def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(description='Train RL agent for LLM attacks')
    parser.add_argument(
        '--config',
        type=str,
        default='configs/rl_training.yaml',
        help='Path to configuration file'
    )
    parser.add_argument(
        '--resume',
        type=str,
        help='Path to checkpoint to resume from'
    )
    parser.add_argument(
        '--device',
        type=str,
        choices=['cpu', 'cuda'],
        help='Device to use for training'
    )
    
    args = parser.parse_args()
    
    # Load configuration
    config = load_config(args.config)
    
    # Override device if specified
    if args.device:
        config['agent']['device'] = args.device
    
    # Create trainer
    trainer = RLTrainer(config)
    
    # Resume from checkpoint if specified
    if args.resume:
        print(f"Resuming from checkpoint: {args.resume}")
        trainer.agent.load(args.resume)
    
    # Start training
    trainer.train()


if __name__ == '__main__':
    main()