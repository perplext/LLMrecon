"""
Multi-Armed Bandit for Provider Optimization in LLMrecon

This module implements various multi-armed bandit algorithms to optimize
provider selection based on success rates and costs.
"""

import numpy as np
from typing import Dict, List, Optional, Tuple, Any
from dataclasses import dataclass
from abc import ABC, abstractmethod
import json
import time
from collections import defaultdict


@dataclass
class ProviderStats:
    """Statistics for a provider/model combination"""
    provider: str
    model: str
    successes: int = 0
    attempts: int = 0
    total_cost: float = 0.0
    total_latency: float = 0.0
    last_updated: float = 0.0
    
    @property
    def success_rate(self) -> float:
        """Calculate success rate"""
        return self.successes / self.attempts if self.attempts > 0 else 0.0
    
    @property
    def avg_cost(self) -> float:
        """Calculate average cost per attempt"""
        return self.total_cost / self.attempts if self.attempts > 0 else 0.0
    
    @property
    def avg_latency(self) -> float:
        """Calculate average latency"""
        return self.total_latency / self.attempts if self.attempts > 0 else 0.0
    
    def update(self, success: bool, cost: float, latency: float):
        """Update statistics with new result"""
        self.attempts += 1
        if success:
            self.successes += 1
        self.total_cost += cost
        self.total_latency += latency
        self.last_updated = time.time()


class BanditAlgorithm(ABC):
    """Base class for bandit algorithms"""
    
    @abstractmethod
    def select_arm(self, arms: List[str], stats: Dict[str, ProviderStats]) -> str:
        """Select an arm (provider) to pull"""
        pass
    
    @abstractmethod
    def update(self, arm: str, reward: float, stats: Dict[str, ProviderStats]):
        """Update algorithm state after observing reward"""
        pass


class EpsilonGreedy(BanditAlgorithm):
    """Epsilon-greedy algorithm with decay"""
    
    def __init__(self, epsilon: float = 0.1, decay: float = 0.99, min_epsilon: float = 0.01):
        self.epsilon = epsilon
        self.decay = decay
        self.min_epsilon = min_epsilon
        
    def select_arm(self, arms: List[str], stats: Dict[str, ProviderStats]) -> str:
        """Select arm using epsilon-greedy strategy"""
        if np.random.random() < self.epsilon:
            # Explore: random selection
            return np.random.choice(arms)
        else:
            # Exploit: select best performing
            best_rate = -1
            best_arm = arms[0]
            
            for arm in arms:
                if arm in stats:
                    rate = stats[arm].success_rate
                    if rate > best_rate:
                        best_rate = rate
                        best_arm = arm
                        
            return best_arm
    
    def update(self, arm: str, reward: float, stats: Dict[str, ProviderStats]):
        """Update epsilon with decay"""
        self.epsilon = max(self.min_epsilon, self.epsilon * self.decay)


class ThompsonSampling(BanditAlgorithm):
    """Thompson Sampling with Beta distribution"""
    
    def __init__(self, alpha: float = 1.0, beta: float = 1.0):
        self.alpha_prior = alpha
        self.beta_prior = beta
        self.arm_params = {}  # Store alpha, beta for each arm
        
    def select_arm(self, arms: List[str], stats: Dict[str, ProviderStats]) -> str:
        """Select arm using Thompson Sampling"""
        samples = {}
        
        for arm in arms:
            # Get or initialize parameters
            if arm not in self.arm_params:
                self.arm_params[arm] = {
                    'alpha': self.alpha_prior,
                    'beta': self.beta_prior
                }
            
            # Sample from Beta distribution
            alpha = self.arm_params[arm]['alpha']
            beta = self.arm_params[arm]['beta']
            samples[arm] = np.random.beta(alpha, beta)
        
        # Select arm with highest sample
        return max(samples, key=samples.get)
    
    def update(self, arm: str, reward: float, stats: Dict[str, ProviderStats]):
        """Update Beta parameters based on reward"""
        if arm not in self.arm_params:
            self.arm_params[arm] = {
                'alpha': self.alpha_prior,
                'beta': self.beta_prior
            }
        
        # Binary reward: success (1) or failure (0)
        if reward > 0.5:  # Threshold for success
            self.arm_params[arm]['alpha'] += 1
        else:
            self.arm_params[arm]['beta'] += 1


class UCB1(BanditAlgorithm):
    """Upper Confidence Bound (UCB1) algorithm"""
    
    def __init__(self, c: float = 2.0):
        self.c = c  # Exploration parameter
        self.total_pulls = 0
        
    def select_arm(self, arms: List[str], stats: Dict[str, ProviderStats]) -> str:
        """Select arm using UCB1 strategy"""
        # Handle unplayed arms first
        for arm in arms:
            if arm not in stats or stats[arm].attempts == 0:
                return arm
        
        # Calculate UCB for each arm
        ucb_values = {}
        for arm in arms:
            if arm in stats:
                avg_reward = stats[arm].success_rate
                n = stats[arm].attempts
                confidence = np.sqrt(self.c * np.log(self.total_pulls) / n)
                ucb_values[arm] = avg_reward + confidence
            else:
                ucb_values[arm] = float('inf')
        
        # Select arm with highest UCB
        return max(ucb_values, key=ucb_values.get)
    
    def update(self, arm: str, reward: float, stats: Dict[str, ProviderStats]):
        """Update total pulls counter"""
        self.total_pulls += 1


class ContextualBandit(BanditAlgorithm):
    """Contextual bandit that considers attack context"""
    
    def __init__(self, feature_dim: int = 10, learning_rate: float = 0.1):
        self.feature_dim = feature_dim
        self.learning_rate = learning_rate
        self.weights = {}  # Weights for each arm
        
    def _extract_context_features(self, context: Dict[str, Any]) -> np.ndarray:
        """Extract features from attack context"""
        # Simple feature extraction (can be enhanced)
        features = np.zeros(self.feature_dim)
        
        # Attack type features
        attack_type = context.get('attack_type', '')
        if 'injection' in attack_type:
            features[0] = 1.0
        elif 'jailbreak' in attack_type:
            features[1] = 1.0
        elif 'manipulation' in attack_type:
            features[2] = 1.0
            
        # Payload features
        payload_length = len(context.get('payload', ''))
        features[3] = min(payload_length / 1000, 1.0)  # Normalized length
        
        # Time features
        hour = context.get('hour', 12)
        features[4] = np.sin(2 * np.pi * hour / 24)
        features[5] = np.cos(2 * np.pi * hour / 24)
        
        # Historical features
        features[6] = context.get('recent_success_rate', 0.5)
        features[7] = context.get('provider_load', 0.5)
        
        return features
    
    def select_arm(self, arms: List[str], stats: Dict[str, ProviderStats], 
                   context: Optional[Dict[str, Any]] = None) -> str:
        """Select arm based on context"""
        if context is None:
            # Fall back to random selection
            return np.random.choice(arms)
        
        # Extract context features
        features = self._extract_context_features(context)
        
        # Calculate expected reward for each arm
        expected_rewards = {}
        for arm in arms:
            if arm not in self.weights:
                # Initialize weights
                self.weights[arm] = np.random.randn(self.feature_dim) * 0.1
            
            # Linear model prediction
            expected_rewards[arm] = np.dot(self.weights[arm], features)
        
        # Add exploration noise (epsilon-greedy style)
        if np.random.random() < 0.1:
            return np.random.choice(arms)
        
        # Select best arm
        return max(expected_rewards, key=expected_rewards.get)
    
    def update(self, arm: str, reward: float, stats: Dict[str, ProviderStats],
               context: Optional[Dict[str, Any]] = None):
        """Update weights using gradient descent"""
        if context is None or arm not in self.weights:
            return
        
        features = self._extract_context_features(context)
        prediction = np.dot(self.weights[arm], features)
        error = reward - prediction
        
        # Gradient descent update
        self.weights[arm] += self.learning_rate * error * features


class MultiArmedBanditOptimizer:
    """
    Main optimizer that manages multiple bandit algorithms for provider selection.
    
    Features:
    - Multiple algorithm support
    - Cost-aware optimization
    - Performance tracking
    - Adaptive strategy selection
    """
    
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        
        # Available providers and models
        self.providers = config.get('providers', {
            'openai': ['gpt-3.5-turbo', 'gpt-4'],
            'anthropic': ['claude-2', 'claude-instant'],
            'google': ['gemini-pro'],
            'meta': ['llama-2-70b']
        })
        
        # Cost per 1K tokens (example values)
        self.costs = config.get('costs', {
            'openai': {'gpt-3.5-turbo': 0.002, 'gpt-4': 0.06},
            'anthropic': {'claude-2': 0.01, 'claude-instant': 0.005},
            'google': {'gemini-pro': 0.001},
            'meta': {'llama-2-70b': 0.0008}
        })
        
        # Initialize statistics
        self.stats = {}
        for provider, models in self.providers.items():
            for model in models:
                arm_id = f"{provider}:{model}"
                self.stats[arm_id] = ProviderStats(provider, model)
        
        # Initialize bandit algorithms
        self.algorithms = {
            'epsilon_greedy': EpsilonGreedy(epsilon=0.2),
            'thompson_sampling': ThompsonSampling(),
            'ucb1': UCB1(c=2.0),
            'contextual': ContextualBandit()
        }
        
        # Current algorithm
        self.current_algorithm = config.get('algorithm', 'thompson_sampling')
        
        # Performance tracking
        self.selection_history = []
        self.reward_history = []
        
    def select_provider(self, 
                       attack_type: str,
                       context: Optional[Dict[str, Any]] = None,
                       budget_constraint: Optional[float] = None) -> Tuple[str, str]:
        """
        Select optimal provider and model for attack.
        
        Args:
            attack_type: Type of attack to perform
            context: Additional context for selection
            budget_constraint: Maximum cost allowed
            
        Returns:
            Tuple of (provider, model)
        """
        # Get available arms considering budget
        available_arms = self._get_available_arms(budget_constraint)
        
        if not available_arms:
            # No providers within budget
            raise ValueError("No providers available within budget constraint")
        
        # Select using current algorithm
        algorithm = self.algorithms[self.current_algorithm]
        
        if isinstance(algorithm, ContextualBandit):
            # Add attack context
            if context is None:
                context = {}
            context['attack_type'] = attack_type
            arm = algorithm.select_arm(available_arms, self.stats, context)
        else:
            arm = algorithm.select_arm(available_arms, self.stats)
        
        # Parse provider and model
        provider, model = arm.split(':')
        
        # Track selection
        self.selection_history.append({
            'timestamp': time.time(),
            'attack_type': attack_type,
            'provider': provider,
            'model': model,
            'algorithm': self.current_algorithm
        })
        
        return provider, model
    
    def update_result(self,
                     provider: str,
                     model: str,
                     success: bool,
                     response_time: float,
                     tokens_used: int,
                     context: Optional[Dict[str, Any]] = None):
        """Update statistics based on attack result"""
        arm_id = f"{provider}:{model}"
        
        # Calculate cost
        cost_per_token = self.costs.get(provider, {}).get(model, 0.001)
        total_cost = (tokens_used / 1000) * cost_per_token
        
        # Update provider stats
        if arm_id in self.stats:
            self.stats[arm_id].update(success, total_cost, response_time)
        
        # Calculate reward (composite metric)
        reward = self._calculate_reward(success, response_time, total_cost)
        
        # Update algorithm
        algorithm = self.algorithms[self.current_algorithm]
        
        if isinstance(algorithm, ContextualBandit):
            algorithm.update(arm_id, reward, self.stats, context)
        else:
            algorithm.update(arm_id, reward, self.stats)
        
        # Track reward
        self.reward_history.append({
            'timestamp': time.time(),
            'arm': arm_id,
            'reward': reward,
            'success': success,
            'cost': total_cost,
            'latency': response_time
        })
    
    def _get_available_arms(self, budget_constraint: Optional[float] = None) -> List[str]:
        """Get arms that satisfy budget constraint"""
        available = []
        
        for provider, models in self.providers.items():
            for model in models:
                arm_id = f"{provider}:{model}"
                
                # Check budget constraint
                if budget_constraint is not None:
                    cost_per_token = self.costs.get(provider, {}).get(model, 0.001)
                    if cost_per_token > budget_constraint:
                        continue
                
                available.append(arm_id)
                
        return available
    
    def _calculate_reward(self, success: bool, response_time: float, cost: float) -> float:
        """
        Calculate composite reward considering multiple factors.
        
        Reward components:
        - Success: Primary factor (0 or 1)
        - Speed: Bonus for fast responses
        - Cost: Penalty for expensive providers
        """
        reward = 0.0
        
        # Success component (50% weight)
        if success:
            reward += 0.5
        
        # Speed component (30% weight)
        # Normalize response time (assume 0-5 seconds range)
        speed_score = max(0, 1 - response_time / 5.0)
        reward += 0.3 * speed_score
        
        # Cost component (20% weight)
        # Normalize cost (assume $0-0.1 range)
        cost_score = max(0, 1 - cost / 0.1)
        reward += 0.2 * cost_score
        
        return reward
    
    def get_statistics(self) -> Dict[str, Any]:
        """Get comprehensive statistics"""
        stats_summary = {}
        
        for arm_id, stats in self.stats.items():
            stats_summary[arm_id] = {
                'attempts': stats.attempts,
                'successes': stats.successes,
                'success_rate': stats.success_rate,
                'avg_cost': stats.avg_cost,
                'avg_latency': stats.avg_latency,
                'total_cost': stats.total_cost
            }
        
        # Overall statistics
        total_attempts = sum(s.attempts for s in self.stats.values())
        total_successes = sum(s.successes for s in self.stats.values())
        total_cost = sum(s.total_cost for s in self.stats.values())
        
        return {
            'provider_stats': stats_summary,
            'total_attempts': total_attempts,
            'total_successes': total_successes,
            'overall_success_rate': total_successes / total_attempts if total_attempts > 0 else 0,
            'total_cost': total_cost,
            'current_algorithm': self.current_algorithm,
            'selection_count': len(self.selection_history),
            'avg_reward': np.mean([r['reward'] for r in self.reward_history]) if self.reward_history else 0
        }
    
    def switch_algorithm(self, algorithm_name: str):
        """Switch to a different bandit algorithm"""
        if algorithm_name not in self.algorithms:
            raise ValueError(f"Unknown algorithm: {algorithm_name}")
        
        self.current_algorithm = algorithm_name
    
    def save_state(self, filepath: str):
        """Save optimizer state to file"""
        state = {
            'stats': {k: v.__dict__ for k, v in self.stats.items()},
            'selection_history': self.selection_history[-1000:],  # Keep last 1000
            'reward_history': self.reward_history[-1000:],
            'current_algorithm': self.current_algorithm,
            'algorithm_states': {}
        }
        
        # Save algorithm-specific state
        for name, algo in self.algorithms.items():
            if hasattr(algo, '__dict__'):
                state['algorithm_states'][name] = algo.__dict__
        
        with open(filepath, 'w') as f:
            json.dump(state, f, indent=2)
    
    def load_state(self, filepath: str):
        """Load optimizer state from file"""
        with open(filepath, 'r') as f:
            state = json.load(f)
        
        # Restore stats
        for arm_id, stats_dict in state['stats'].items():
            self.stats[arm_id] = ProviderStats(**stats_dict)
        
        # Restore history
        self.selection_history = state['selection_history']
        self.reward_history = state['reward_history']
        self.current_algorithm = state['current_algorithm']
        
        # Restore algorithm states
        for name, algo_state in state.get('algorithm_states', {}).items():
            if name in self.algorithms:
                for key, value in algo_state.items():
                    setattr(self.algorithms[name], key, value)