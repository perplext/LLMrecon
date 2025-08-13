"""
LLM Attack Environment for Reinforcement Learning

This module implements an OpenAI Gym-compatible environment for training
RL agents to optimize LLM attack strategies.
"""

import gym
import numpy as np
from gym import spaces
from typing import Dict, Tuple, Any, Optional
import json
import hashlib
from dataclasses import dataclass
from enum import Enum


class AttackType(Enum):
    """Types of attacks available to the agent"""
    PROMPT_INJECTION = 0
    JAILBREAK = 1
    CONTEXT_MANIPULATION = 2
    TOKEN_SMUGGLING = 3
    ROLE_PLAY = 4


class TargetModel(Enum):
    """Target LLM models"""
    GPT_35_TURBO = 0
    GPT_4 = 1
    CLAUDE_2 = 2
    LLAMA_2 = 3
    GEMINI = 4


@dataclass
class AttackResult:
    """Result of an attack attempt"""
    success: bool
    response: str
    tokens_used: int
    time_taken: float
    detection_score: float  # 0-1, higher means more likely detected


class LLMAttackEnv(gym.Env):
    """
    OpenAI Gym environment for LLM attack optimization.
    
    The agent learns to select optimal attack strategies based on:
    - Target model characteristics
    - Historical attack outcomes
    - Current context
    """
    
    metadata = {'render.modes': ['human', 'json']}
    
    def __init__(self, config: Dict[str, Any]):
        super().__init__()
        
        self.config = config
        self.max_steps = config.get('max_steps', 100)
        self.current_step = 0
        self.attack_history = []
        self.total_reward = 0.0
        
        # Define action space
        self.action_space = spaces.Dict({
            'attack_type': spaces.Discrete(len(AttackType)),
            'intensity': spaces.Box(low=0.0, high=1.0, shape=(1,), dtype=np.float32),
            'obfuscation': spaces.Box(low=0.0, high=1.0, shape=(1,), dtype=np.float32),
            'target_model': spaces.Discrete(len(TargetModel)),
            'technique_params': spaces.Box(low=0.0, high=1.0, shape=(5,), dtype=np.float32)
        })
        
        # Define observation space
        self.observation_space = spaces.Dict({
            # Current model state (embedding of model's recent responses)
            'model_state': spaces.Box(low=-1.0, high=1.0, shape=(128,), dtype=np.float32),
            
            # Attack history features (last 10 attacks)
            'history_features': spaces.Box(low=0.0, high=1.0, shape=(10, 8), dtype=np.float32),
            
            # Context embedding (current conversation context)
            'context': spaces.Box(low=-1.0, high=1.0, shape=(256,), dtype=np.float32),
            
            # Model-specific features
            'model_features': spaces.Box(low=0.0, high=1.0, shape=(16,), dtype=np.float32),
            
            # Time and resource constraints
            'constraints': spaces.Box(low=0.0, high=1.0, shape=(4,), dtype=np.float32)
        })
        
        # Initialize state
        self._reset_state()
        
    def _reset_state(self):
        """Reset internal state"""
        self.current_step = 0
        self.attack_history = []
        self.total_reward = 0.0
        self.model_state = np.zeros(128, dtype=np.float32)
        self.context = np.zeros(256, dtype=np.float32)
        
    def reset(self) -> Dict[str, np.ndarray]:
        """Reset the environment to initial state"""
        self._reset_state()
        return self._get_observation()
    
    def step(self, action: Dict[str, np.ndarray]) -> Tuple[Dict[str, np.ndarray], float, bool, Dict[str, Any]]:
        """
        Execute an attack action and return the results.
        
        Args:
            action: Dictionary containing attack parameters
            
        Returns:
            observation: New state after attack
            reward: Reward for the action
            done: Whether episode is complete
            info: Additional information
        """
        self.current_step += 1
        
        # Execute attack (simulated for now)
        attack_result = self._execute_attack(action)
        
        # Calculate reward
        reward = self._calculate_reward(attack_result, action)
        self.total_reward += reward
        
        # Update state based on attack outcome
        self._update_state(attack_result, action)
        
        # Check if episode is done
        done = self._is_done()
        
        # Get new observation
        observation = self._get_observation()
        
        # Additional info for logging
        info = {
            'attack_result': attack_result,
            'total_reward': self.total_reward,
            'step': self.current_step,
            'success_rate': self._calculate_success_rate()
        }
        
        return observation, reward, done, info
    
    def _execute_attack(self, action: Dict[str, np.ndarray]) -> AttackResult:
        """
        Execute the attack (simulated for now, will integrate with real API).
        
        In production, this will:
        1. Construct attack payload based on action
        2. Send to target LLM
        3. Analyze response
        4. Return results
        """
        # Extract action parameters
        attack_type = AttackType(int(action['attack_type']))
        intensity = float(action['intensity'][0])
        obfuscation = float(action['obfuscation'][0])
        target_model = TargetModel(int(action['target_model']))
        
        # Simulate attack outcome based on parameters
        # In reality, this would call the actual LLM API
        base_success_rate = 0.3
        
        # Adjust success rate based on attack parameters
        if attack_type == AttackType.PROMPT_INJECTION:
            base_success_rate += 0.1 * intensity
        elif attack_type == AttackType.JAILBREAK:
            base_success_rate += 0.15 * intensity * (1 - obfuscation * 0.5)
        
        # Model-specific adjustments
        if target_model == TargetModel.GPT_35_TURBO:
            base_success_rate += 0.05
        elif target_model == TargetModel.GPT_4:
            base_success_rate -= 0.1
            
        # Add randomness
        success = np.random.random() < base_success_rate
        
        # Simulate other metrics
        tokens_used = int(100 + 200 * intensity)
        time_taken = 0.5 + 2.0 * intensity
        detection_score = min(1.0, obfuscation * 0.3 + intensity * 0.5)
        
        return AttackResult(
            success=success,
            response="Simulated response",
            tokens_used=tokens_used,
            time_taken=time_taken,
            detection_score=detection_score
        )
    
    def _calculate_reward(self, result: AttackResult, action: Dict[str, np.ndarray]) -> float:
        """
        Calculate reward based on attack outcome.
        
        Reward components:
        - Success: +10 for successful attack
        - Efficiency: Bonus for using fewer tokens/time
        - Stealth: Bonus for low detection score
        - Penalty: For failed attacks and high resource usage
        """
        reward = 0.0
        
        # Success reward
        if result.success:
            reward += 10.0
        else:
            reward -= 2.0
            
        # Efficiency bonus (fewer tokens is better)
        token_efficiency = max(0, 1 - result.tokens_used / 500)
        reward += 2.0 * token_efficiency
        
        # Time efficiency
        time_efficiency = max(0, 1 - result.time_taken / 5.0)
        reward += 1.0 * time_efficiency
        
        # Stealth bonus (lower detection is better)
        stealth_bonus = 1 - result.detection_score
        reward += 3.0 * stealth_bonus
        
        # Exploration bonus for trying different techniques
        intensity = float(action['intensity'][0])
        if 0.3 < intensity < 0.7:  # Moderate intensity
            reward += 0.5
            
        return reward
    
    def _update_state(self, result: AttackResult, action: Dict[str, np.ndarray]):
        """Update internal state based on attack outcome"""
        # Update model state (simplified - in practice, use actual embeddings)
        self.model_state = np.random.randn(128).astype(np.float32) * 0.1 + self.model_state * 0.9
        
        # Update context
        self.context = np.random.randn(256).astype(np.float32) * 0.1 + self.context * 0.9
        
        # Add to history
        history_entry = {
            'attack_type': int(action['attack_type']),
            'success': result.success,
            'intensity': float(action['intensity'][0]),
            'tokens_used': result.tokens_used,
            'detection_score': result.detection_score
        }
        self.attack_history.append(history_entry)
        
        # Keep only last 10 attacks
        if len(self.attack_history) > 10:
            self.attack_history.pop(0)
    
    def _get_observation(self) -> Dict[str, np.ndarray]:
        """Get current observation for the agent"""
        # Prepare history features
        history_features = np.zeros((10, 8), dtype=np.float32)
        for i, entry in enumerate(self.attack_history[-10:]):
            history_features[i] = [
                entry['attack_type'] / 5.0,  # Normalize
                float(entry['success']),
                entry['intensity'],
                entry['tokens_used'] / 500.0,  # Normalize
                entry['detection_score'],
                0.0,  # Placeholder for additional features
                0.0,
                0.0
            ]
        
        # Model-specific features (simplified)
        model_features = np.random.rand(16).astype(np.float32)
        
        # Constraints (remaining steps, token budget, etc.)
        constraints = np.array([
            self.current_step / self.max_steps,
            0.8,  # Token budget remaining
            0.5,  # Time budget remaining
            0.2   # Detection threshold
        ], dtype=np.float32)
        
        return {
            'model_state': self.model_state,
            'history_features': history_features,
            'context': self.context,
            'model_features': model_features,
            'constraints': constraints
        }
    
    def _is_done(self) -> bool:
        """Check if episode should end"""
        # End if max steps reached
        if self.current_step >= self.max_steps:
            return True
            
        # End if success rate is very high (early stopping)
        if len(self.attack_history) >= 10:
            recent_success = sum(h['success'] for h in self.attack_history[-10:]) / 10
            if recent_success >= 0.9:
                return True
                
        return False
    
    def _calculate_success_rate(self) -> float:
        """Calculate overall success rate"""
        if not self.attack_history:
            return 0.0
        successful = sum(h['success'] for h in self.attack_history)
        return successful / len(self.attack_history)
    
    def render(self, mode='human'):
        """Render the environment state"""
        if mode == 'human':
            print(f"\n=== LLM Attack Environment ===")
            print(f"Step: {self.current_step}/{self.max_steps}")
            print(f"Total Reward: {self.total_reward:.2f}")
            print(f"Success Rate: {self._calculate_success_rate():.2%}")
            if self.attack_history:
                last_attack = self.attack_history[-1]
                print(f"Last Attack: Type={AttackType(last_attack['attack_type']).name}, "
                      f"Success={last_attack['success']}")
        elif mode == 'json':
            return json.dumps({
                'step': self.current_step,
                'total_reward': self.total_reward,
                'success_rate': self._calculate_success_rate(),
                'history': self.attack_history[-5:]  # Last 5 attacks
            }, indent=2)
    
    def close(self):
        """Clean up resources"""
        pass


# Factory function for creating environments
def make_attack_env(config: Optional[Dict[str, Any]] = None) -> LLMAttackEnv:
    """Create an LLM attack environment with given configuration"""
    default_config = {
        'max_steps': 100,
        'reward_shaping': True,
        'use_real_api': False  # Set to True when integrating with actual LLM APIs
    }
    
    if config:
        default_config.update(config)
        
    return LLMAttackEnv(default_config)