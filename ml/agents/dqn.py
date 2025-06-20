"""
Deep Q-Network (DQN) Agent for LLM Attack Optimization

This module implements a DQN agent with prioritized experience replay
for learning optimal attack strategies.
"""

import torch
import torch.nn as nn
import torch.optim as optim
import torch.nn.functional as F
import numpy as np
from typing import Dict, Tuple, Optional, List
from collections import deque, namedtuple
import random
from dataclasses import dataclass


# Experience tuple for replay buffer
Experience = namedtuple('Experience', 
    ['state', 'action', 'reward', 'next_state', 'done'])


@dataclass
class DQNConfig:
    """Configuration for DQN agent"""
    # Network architecture
    hidden_sizes: List[int] = (256, 256, 128)
    
    # Training hyperparameters
    learning_rate: float = 1e-4
    batch_size: int = 32
    gamma: float = 0.99
    tau: float = 0.005  # Soft update parameter
    
    # Exploration
    epsilon_start: float = 1.0
    epsilon_end: float = 0.01
    epsilon_decay: float = 0.995
    
    # Experience replay
    buffer_size: int = 10000
    prioritized_replay: bool = True
    alpha: float = 0.6  # Prioritization exponent
    beta_start: float = 0.4  # Importance sampling exponent
    beta_end: float = 1.0
    
    # Device
    device: str = "cuda" if torch.cuda.is_available() else "cpu"


class AttackQNetwork(nn.Module):
    """
    Neural network for approximating Q-values for attack actions.
    
    Takes flattened state as input and outputs Q-values for each
    discrete action combination.
    """
    
    def __init__(self, state_size: int, action_size: int, hidden_sizes: Tuple[int]):
        super().__init__()
        
        # Build network layers
        layers = []
        prev_size = state_size
        
        for hidden_size in hidden_sizes:
            layers.extend([
                nn.Linear(prev_size, hidden_size),
                nn.ReLU(),
                nn.BatchNorm1d(hidden_size),
                nn.Dropout(0.1)
            ])
            prev_size = hidden_size
            
        # Output layer
        layers.append(nn.Linear(prev_size, action_size))
        
        self.network = nn.Sequential(*layers)
        
    def forward(self, state: torch.Tensor) -> torch.Tensor:
        """Forward pass through the network"""
        return self.network(state)


class PrioritizedReplayBuffer:
    """
    Prioritized experience replay buffer.
    
    Samples experiences based on their TD error (priority).
    """
    
    def __init__(self, capacity: int, alpha: float = 0.6):
        self.capacity = capacity
        self.alpha = alpha
        self.buffer = []
        self.priorities = np.zeros(capacity, dtype=np.float32)
        self.position = 0
        self.max_priority = 1.0
        
    def push(self, experience: Experience):
        """Add experience to buffer with max priority"""
        if len(self.buffer) < self.capacity:
            self.buffer.append(experience)
        else:
            self.buffer[self.position] = experience
            
        # New experiences get max priority
        self.priorities[self.position] = self.max_priority ** self.alpha
        self.position = (self.position + 1) % self.capacity
        
    def sample(self, batch_size: int, beta: float = 0.4) -> Tuple[List[Experience], np.ndarray, np.ndarray]:
        """
        Sample batch of experiences based on priorities.
        
        Returns:
            experiences: Batch of experiences
            weights: Importance sampling weights
            indices: Indices of sampled experiences
        """
        if len(self.buffer) == 0:
            return [], np.array([]), np.array([])
            
        # Calculate sampling probabilities
        priorities = self.priorities[:len(self.buffer)]
        probs = priorities / priorities.sum()
        
        # Sample indices
        indices = np.random.choice(len(self.buffer), batch_size, p=probs)
        experiences = [self.buffer[idx] for idx in indices]
        
        # Calculate importance sampling weights
        weights = (len(self.buffer) * probs[indices]) ** (-beta)
        weights /= weights.max()  # Normalize
        
        return experiences, weights, indices
    
    def update_priorities(self, indices: np.ndarray, td_errors: np.ndarray):
        """Update priorities based on TD errors"""
        priorities = (np.abs(td_errors) + 1e-6) ** self.alpha
        self.priorities[indices] = priorities
        self.max_priority = max(self.max_priority, priorities.max())
        
    def __len__(self):
        return len(self.buffer)


class DQNAgent:
    """
    DQN agent for learning optimal LLM attack strategies.
    
    Features:
    - Deep Q-learning with target network
    - Prioritized experience replay
    - Epsilon-greedy exploration
    - Soft target network updates
    """
    
    def __init__(self, state_space: Dict, action_space: Dict, config: Optional[DQNConfig] = None):
        self.config = config or DQNConfig()
        self.device = torch.device(self.config.device)
        
        # Calculate flattened sizes
        self.state_size = self._calculate_state_size(state_space)
        self.action_size = self._calculate_action_size(action_space)
        
        # Store action space for decoding
        self.action_space = action_space
        
        # Initialize networks
        self.q_network = AttackQNetwork(
            self.state_size, 
            self.action_size,
            self.config.hidden_sizes
        ).to(self.device)
        
        self.target_network = AttackQNetwork(
            self.state_size,
            self.action_size,
            self.config.hidden_sizes
        ).to(self.device)
        
        # Initialize target network with same weights
        self.target_network.load_state_dict(self.q_network.state_dict())
        
        # Optimizer
        self.optimizer = optim.Adam(
            self.q_network.parameters(), 
            lr=self.config.learning_rate
        )
        
        # Replay buffer
        self.memory = PrioritizedReplayBuffer(
            self.config.buffer_size,
            self.config.alpha
        )
        
        # Training state
        self.epsilon = self.config.epsilon_start
        self.beta = self.config.beta_start
        self.training_step = 0
        
    def _calculate_state_size(self, state_space: Dict) -> int:
        """Calculate total size of flattened state"""
        total_size = 0
        for key, space in state_space.spaces.items():
            total_size += np.prod(space.shape)
        return int(total_size)
    
    def _calculate_action_size(self, action_space: Dict) -> int:
        """
        Calculate total number of discrete action combinations.
        
        For now, we discretize continuous actions into bins.
        """
        # Simplified: treat each component as discrete choices
        total_actions = 1
        for key, space in action_space.spaces.items():
            if hasattr(space, 'n'):
                total_actions *= space.n
            else:
                # Discretize continuous actions into 5 bins
                total_actions *= 5
        return int(total_actions)
    
    def _flatten_state(self, state: Dict[str, np.ndarray]) -> torch.Tensor:
        """Flatten state dictionary into single tensor"""
        flat_state = []
        for key in sorted(state.keys()):
            flat_state.append(state[key].flatten())
        
        flat_array = np.concatenate(flat_state)
        return torch.FloatTensor(flat_array).unsqueeze(0).to(self.device)
    
    def _decode_action(self, action_idx: int) -> Dict[str, np.ndarray]:
        """
        Decode discrete action index back to action dictionary.
        
        This is simplified - in practice, we'd use more sophisticated
        action encoding/decoding.
        """
        # For now, use simple modulo arithmetic to decode
        actions = {}
        
        # Attack type (5 choices)
        actions['attack_type'] = np.array(action_idx % 5)
        action_idx //= 5
        
        # Intensity (5 discrete levels)
        intensity_level = action_idx % 5
        actions['intensity'] = np.array([intensity_level / 4.0])
        action_idx //= 5
        
        # Obfuscation (5 discrete levels)
        obfuscation_level = action_idx % 5
        actions['obfuscation'] = np.array([obfuscation_level / 4.0])
        action_idx //= 5
        
        # Target model (5 choices)
        actions['target_model'] = np.array(action_idx % 5)
        
        # Technique params (simplified)
        actions['technique_params'] = np.random.rand(5).astype(np.float32)
        
        return actions
    
    def act(self, state: Dict[str, np.ndarray], explore: bool = True) -> Dict[str, np.ndarray]:
        """
        Select action using epsilon-greedy policy.
        
        Args:
            state: Current environment state
            explore: Whether to use exploration
            
        Returns:
            action: Selected action
        """
        # Epsilon-greedy exploration
        if explore and random.random() < self.epsilon:
            # Random action
            action_idx = random.randrange(self.action_size)
        else:
            # Greedy action
            state_tensor = self._flatten_state(state)
            
            with torch.no_grad():
                q_values = self.q_network(state_tensor)
                action_idx = q_values.argmax().item()
                
        return self._decode_action(action_idx)
    
    def remember(self, state: Dict, action: Dict, reward: float, 
                 next_state: Dict, done: bool):
        """Store experience in replay buffer"""
        experience = Experience(state, action, reward, next_state, done)
        self.memory.push(experience)
        
    def learn(self):
        """
        Update Q-network using experiences from replay buffer.
        
        Returns:
            loss: Training loss (if training occurred)
        """
        if len(self.memory) < self.config.batch_size:
            return None
            
        # Sample batch
        experiences, weights, indices = self.memory.sample(
            self.config.batch_size, 
            self.beta
        )
        
        # Convert to tensors
        states = torch.cat([self._flatten_state(e.state) for e in experiences])
        next_states = torch.cat([self._flatten_state(e.next_state) for e in experiences])
        rewards = torch.FloatTensor([e.reward for e in experiences]).to(self.device)
        dones = torch.FloatTensor([e.done for e in experiences]).to(self.device)
        weights = torch.FloatTensor(weights).to(self.device)
        
        # Encode actions to indices
        action_indices = []
        for e in experiences:
            # Simplified encoding - inverse of _decode_action
            idx = 0
            idx += int(e.action['attack_type'])
            idx += int(e.action['intensity'][0] * 4) * 5
            idx += int(e.action['obfuscation'][0] * 4) * 25
            idx += int(e.action['target_model']) * 125
            action_indices.append(idx)
        
        action_indices = torch.LongTensor(action_indices).to(self.device)
        
        # Current Q values
        current_q_values = self.q_network(states).gather(1, action_indices.unsqueeze(1))
        
        # Next Q values from target network
        with torch.no_grad():
            next_q_values = self.target_network(next_states).max(1)[0]
            target_q_values = rewards + (self.config.gamma * next_q_values * (1 - dones))
            
        # TD errors for priority updates
        td_errors = (current_q_values.squeeze() - target_q_values).detach().cpu().numpy()
        self.memory.update_priorities(indices, td_errors)
        
        # Weighted MSE loss
        loss = (weights * F.mse_loss(current_q_values.squeeze(), target_q_values, reduction='none')).mean()
        
        # Optimize
        self.optimizer.zero_grad()
        loss.backward()
        torch.nn.utils.clip_grad_norm_(self.q_network.parameters(), 1.0)
        self.optimizer.step()
        
        # Soft update target network
        self._soft_update()
        
        # Update exploration parameters
        self._update_exploration()
        
        self.training_step += 1
        
        return loss.item()
    
    def _soft_update(self):
        """Soft update of target network parameters"""
        for target_param, local_param in zip(
            self.target_network.parameters(), 
            self.q_network.parameters()
        ):
            target_param.data.copy_(
                self.config.tau * local_param.data + 
                (1.0 - self.config.tau) * target_param.data
            )
            
    def _update_exploration(self):
        """Update exploration parameters"""
        # Decay epsilon
        self.epsilon = max(
            self.config.epsilon_end,
            self.epsilon * self.config.epsilon_decay
        )
        
        # Increase beta (importance sampling)
        progress = min(1.0, self.training_step / 10000)
        self.beta = self.config.beta_start + progress * (self.config.beta_end - self.config.beta_start)
        
    def save(self, filepath: str):
        """Save agent state"""
        torch.save({
            'q_network_state_dict': self.q_network.state_dict(),
            'target_network_state_dict': self.target_network.state_dict(),
            'optimizer_state_dict': self.optimizer.state_dict(),
            'epsilon': self.epsilon,
            'beta': self.beta,
            'training_step': self.training_step
        }, filepath)
        
    def load(self, filepath: str):
        """Load agent state"""
        checkpoint = torch.load(filepath, map_location=self.device)
        self.q_network.load_state_dict(checkpoint['q_network_state_dict'])
        self.target_network.load_state_dict(checkpoint['target_network_state_dict'])
        self.optimizer.load_state_dict(checkpoint['optimizer_state_dict'])
        self.epsilon = checkpoint['epsilon']
        self.beta = checkpoint['beta']
        self.training_step = checkpoint['training_step']