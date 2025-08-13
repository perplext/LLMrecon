"""
ML Model Storage Demo for LLMrecon

This script demonstrates how to use the model storage infrastructure
with the ML components (DQN, Multi-Armed Bandit, etc.).
"""

import os
import sys
import json
import numpy as np
from datetime import datetime

# Add parent directories to path
sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))))

from ml.agents.dqn import DQNAgent, DQNConfig
from ml.agents.multi_armed_bandit import MultiArmedBanditOptimizer
from ml.storage.model_storage import create_model_registry, ModelMetadata
from ml.storage.model_versioning import ModelVersionManager, ModelStage


def demo_dqn_storage():
    """Demonstrate storing and loading DQN models"""
    print("\n=== DQN Model Storage Demo ===")
    
    # Create model registry
    registry = create_model_registry({
        'primary': 'local',
        'local': {
            'enabled': True,
            'path': 'ml/models/demo'
        }
    })
    
    # Create version manager
    version_manager = ModelVersionManager(registry)
    
    # Create and train a DQN agent
    state_space = {
        'type': 'dict',
        'spaces': {
            'payload_length': {'type': 'box', 'low': 0, 'high': 1000},
            'obfuscation': {'type': 'box', 'low': 0, 'high': 1},
            'model_confidence': {'type': 'box', 'low': 0, 'high': 1}
        }
    }
    
    action_space = {
        'type': 'discrete',
        'n': 5  # 5 different attack types
    }
    
    config = DQNConfig(
        learning_rate=0.001,
        buffer_size=10000,
        batch_size=32,
        epsilon=0.1
    )
    
    agent = DQNAgent(state_space, action_space, config)
    
    # Simulate some training
    print("Simulating training...")
    for i in range(100):
        # Fake state
        state = {
            'payload_length': np.random.randint(0, 1000),
            'obfuscation': np.random.random(),
            'model_confidence': np.random.random()
        }
        
        # Get action
        action = agent.act(state)
        
        # Fake reward
        reward = np.random.random()
        
        # Fake next state
        next_state = {
            'payload_length': np.random.randint(0, 1000),
            'obfuscation': np.random.random(),
            'model_confidence': np.random.random()
        }
        
        # Store experience
        agent.remember(state, action, reward, next_state, done=False)
        
        # Train periodically
        if i % 10 == 0 and len(agent.memory) > agent.config.batch_size:
            loss = agent.replay()
    
    # Calculate fake performance metrics
    performance_metrics = {
        'success_rate': 0.75,
        'avg_reward': 0.65,
        'training_episodes': 100,
        'final_epsilon': agent.epsilon
    }
    
    # Save the model
    print("\nSaving DQN model...")
    metadata = version_manager.register_model(
        model=agent,
        name="attack_optimizer",
        algorithm="dqn",
        model_type="custom",
        tags=["reinforcement-learning", "attack-optimization"],
        description="DQN agent for optimizing attack strategies",
        training_params={
            'learning_rate': config.learning_rate,
            'buffer_size': config.buffer_size,
            'batch_size': config.batch_size,
            'training_episodes': 100
        },
        performance_metrics=performance_metrics,
        author="LLMrecon ML Team"
    )
    
    print(f"Model saved: {metadata.model_id} v{metadata.version}")
    
    # Transition to staging
    print("\nTransitioning model to staging...")
    version_manager.transition_stage(
        metadata.model_id,
        metadata.version,
        ModelStage.STAGING
    )
    
    # Load the model back
    print("\nLoading model from storage...")
    loaded_agent = registry.load_model(metadata.model_id, metadata.version)
    
    # Verify it works
    test_state = {
        'payload_length': 500,
        'obfuscation': 0.5,
        'model_confidence': 0.8
    }
    action = loaded_agent.act(test_state)
    print(f"Loaded model action on test state: {action}")
    
    # Create an improved version
    print("\nCreating improved version...")
    
    # Simulate more training
    for i in range(50):
        state = {
            'payload_length': np.random.randint(0, 1000),
            'obfuscation': np.random.random(),
            'model_confidence': np.random.random()
        }
        action = loaded_agent.act(state)
        reward = np.random.random() * 1.2  # Better rewards
        next_state = {
            'payload_length': np.random.randint(0, 1000),
            'obfuscation': np.random.random(),
            'model_confidence': np.random.random()
        }
        loaded_agent.remember(state, action, reward, next_state, done=False)
        
        if i % 10 == 0 and len(loaded_agent.memory) > loaded_agent.config.batch_size:
            loaded_agent.replay()
    
    # Save improved version
    improved_metrics = {
        'success_rate': 0.82,  # Improved!
        'avg_reward': 0.78,
        'training_episodes': 150,
        'final_epsilon': loaded_agent.epsilon
    }
    
    improved_metadata = version_manager.register_model(
        model=loaded_agent,
        name="attack_optimizer",
        algorithm="dqn",
        model_type="custom",
        parent_model=(metadata.model_id, metadata.version),
        bump_type="minor",  # Minor version bump
        tags=["reinforcement-learning", "attack-optimization", "improved"],
        description="Improved DQN agent with additional training",
        training_params={
            'learning_rate': config.learning_rate,
            'buffer_size': config.buffer_size,
            'batch_size': config.batch_size,
            'training_episodes': 150,
            'parent_model': f"{metadata.model_id}:{metadata.version}"
        },
        performance_metrics=improved_metrics,
        created_from="fine-tuning"
    )
    
    print(f"Improved model saved: {improved_metadata.model_id} v{improved_metadata.version}")
    
    # Compare versions
    print("\nComparing model versions...")
    comparison = version_manager.compare_versions(
        metadata.model_id,
        metadata.version,
        improved_metadata.version,
        metric='success_rate'
    )
    
    print(f"Version comparison:")
    print(f"  v{metadata.version}: {comparison['version1']['metrics'].get('success_rate', 0):.2%}")
    print(f"  v{improved_metadata.version}: {comparison['version2']['metrics'].get('success_rate', 0):.2%}")
    print(f"  Improvement: {comparison['improvement']['change']:.1f}%")
    
    # Promote to production
    print("\nPromoting improved model to production...")
    version_manager.transition_stage(
        improved_metadata.model_id,
        improved_metadata.version,
        ModelStage.PRODUCTION
    )
    
    # Get production model
    print("\nLoading production model...")
    prod_model = version_manager.get_production_model(metadata.model_id)
    print(f"Production model loaded successfully")
    
    # Show lineage
    print("\nModel lineage:")
    lineage = version_manager.get_lineage(
        improved_metadata.model_id,
        improved_metadata.version
    )
    print(json.dumps(lineage, indent=2))


def demo_bandit_storage():
    """Demonstrate storing Multi-Armed Bandit models"""
    print("\n\n=== Multi-Armed Bandit Storage Demo ===")
    
    # Create model registry
    registry = create_model_registry()
    version_manager = ModelVersionManager(registry)
    
    # Create bandit optimizer
    config = {
        'providers': {
            'openai': ['gpt-3.5-turbo', 'gpt-4'],
            'anthropic': ['claude-2', 'claude-instant']
        },
        'algorithm': 'thompson_sampling'
    }
    
    bandit = MultiArmedBanditOptimizer(config)
    
    # Simulate some usage
    print("Simulating bandit usage...")
    for i in range(50):
        provider, model = bandit.select_provider(
            attack_type="prompt_injection",
            context={'hour': 14, 'recent_success_rate': 0.7}
        )
        
        # Simulate result
        success = np.random.random() > 0.3
        response_time = np.random.uniform(0.5, 3.0)
        tokens_used = np.random.randint(100, 500)
        
        bandit.update_result(
            provider, model, success, response_time, tokens_used
        )
    
    # Get statistics
    stats = bandit.get_statistics()
    
    # Save the bandit model
    print("\nSaving bandit model...")
    metadata = version_manager.register_model(
        model=bandit,
        name="provider_optimizer",
        algorithm="multi_armed_bandit",
        model_type="custom",
        tags=["optimization", "provider-selection", "thompson-sampling"],
        description="Multi-armed bandit for optimal provider selection",
        training_params={
            'algorithm': 'thompson_sampling',
            'providers': list(config['providers'].keys()),
            'total_pulls': stats['total_attempts']
        },
        performance_metrics={
            'overall_success_rate': stats['overall_success_rate'],
            'avg_reward': stats['avg_reward'],
            'total_cost': stats['total_cost']
        }
    )
    
    print(f"Bandit model saved: {metadata.model_id} v{metadata.version}")
    
    # List all models
    print("\n\nListing all saved models:")
    all_models = registry.list_models()
    for model in all_models:
        stage = version_manager.stages.get(
            f"{model.model_id}:{model.version}",
            ModelStage.DEVELOPMENT
        )
        print(f"  - {model.model_id} v{model.version} ({model.algorithm}) - {stage.value}")
    
    # Export model history
    print("\n\nExporting model history:")
    history = version_manager.export_model_history("dqn_attack_optimizer")
    print(f"Total versions: {history['total_versions']}")
    for ver in history['versions']:
        print(f"  - v{ver['version']} ({ver['stage']}) - Success rate: {ver['performance_metrics'].get('success_rate', 'N/A')}")


def demo_cleanup():
    """Demonstrate model cleanup"""
    print("\n\n=== Model Cleanup Demo ===")
    
    registry = create_model_registry()
    version_manager = ModelVersionManager(registry)
    
    # Create multiple versions for cleanup demo
    print("Creating multiple model versions...")
    for i in range(8):
        # Create dummy model
        dummy_model = {'version': i, 'data': np.random.randn(10, 10)}
        
        metadata = version_manager.register_model(
            model=dummy_model,
            name="cleanup_test",
            algorithm="test",
            model_type="custom",
            bump_type="patch" if i > 0 else None,
            parent_model=("test_cleanup_test", f"1.0.{i-1}") if i > 0 else None,
            performance_metrics={'accuracy': 0.5 + i * 0.05}
        )
        
        # Promote some to different stages
        if i == 3:
            version_manager.transition_stage(
                metadata.model_id,
                metadata.version,
                ModelStage.STAGING
            )
        elif i == 5:
            version_manager.transition_stage(
                metadata.model_id,
                metadata.version,
                ModelStage.STAGING
            )
            version_manager.transition_stage(
                metadata.model_id,
                metadata.version,
                ModelStage.PRODUCTION
            )
    
    # List versions before cleanup
    print("\nVersions before cleanup:")
    models = registry.list_models(algorithm="test")
    for model in models:
        stage = version_manager.stages.get(
            f"{model.model_id}:{model.version}",
            ModelStage.DEVELOPMENT
        )
        print(f"  - v{model.version} - {stage.value}")
    
    # Perform cleanup
    print("\nPerforming cleanup (keep last 3 + production/staging)...")
    deleted = version_manager.cleanup_old_versions(
        "test_cleanup_test",
        keep_last_n=3,
        keep_production=True,
        keep_staging=True
    )
    
    print(f"Deleted versions: {deleted}")
    
    # List versions after cleanup
    print("\nVersions after cleanup:")
    models = registry.list_models(algorithm="test")
    for model in models:
        stage = version_manager.stages.get(
            f"{model.model_id}:{model.version}",
            ModelStage.DEVELOPMENT
        )
        print(f"  - v{model.version} - {stage.value}")


if __name__ == "__main__":
    # Create necessary directories
    os.makedirs("ml/models/demo", exist_ok=True)
    os.makedirs("ml/models", exist_ok=True)
    
    # Run demos
    demo_dqn_storage()
    demo_bandit_storage()
    demo_cleanup()
    
    print("\n\n=== Storage Demo Complete ===")
    print("Check ml/models/ directory for stored models and metadata")