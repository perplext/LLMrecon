"""
Test script to verify ML environment setup.

Run this to ensure all components are working correctly.
"""

import sys
import os
import torch
import numpy as np
from pathlib import Path

# Add parent directory to path
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

def test_gpu():
    """Test GPU availability and basic operations"""
    print("=== GPU Test ===")
    
    # Check PyTorch CUDA
    cuda_available = torch.cuda.is_available()
    print(f"PyTorch CUDA available: {cuda_available}")
    
    if cuda_available:
        print(f"CUDA device count: {torch.cuda.device_count()}")
        print(f"Current device: {torch.cuda.current_device()}")
        print(f"Device name: {torch.cuda.get_device_name(0)}")
        
        # Test computation
        x = torch.randn(1000, 1000).cuda()
        y = torch.randn(1000, 1000).cuda()
        z = torch.matmul(x, y)
        print(f"GPU computation test: SUCCESS (output shape: {z.shape})")
    else:
        print("Running on CPU")
    
    print()
    return cuda_available


def test_environment():
    """Test RL environment"""
    print("=== Environment Test ===")
    
    try:
        from ml.environments.attack_env import make_attack_env
        
        # Create environment
        env = make_attack_env()
        print(f"Environment created: {type(env).__name__}")
        print(f"Observation space: {env.observation_space}")
        print(f"Action space: {env.action_space}")
        
        # Test reset
        obs = env.reset()
        print(f"Reset successful. Observation keys: {list(obs.keys())}")
        
        # Test step
        action = env.action_space.sample()
        obs, reward, done, info = env.step(action)
        print(f"Step successful. Reward: {reward:.2f}")
        
        print("Environment test: SUCCESS")
    except Exception as e:
        print(f"Environment test: FAILED - {e}")
        return False
    
    print()
    return True


def test_agent():
    """Test DQN agent"""
    print("=== Agent Test ===")
    
    try:
        from ml.environments.attack_env import make_attack_env
        from ml.agents.dqn import DQNAgent, DQNConfig
        
        # Create environment and agent
        env = make_attack_env()
        config = DQNConfig(device="cuda" if torch.cuda.is_available() else "cpu")
        agent = DQNAgent(env.observation_space, env.action_space, config)
        
        print(f"Agent created: {type(agent).__name__}")
        print(f"Q-network parameters: {sum(p.numel() for p in agent.q_network.parameters()):,}")
        
        # Test action selection
        obs = env.reset()
        action = agent.act(obs, explore=True)
        print(f"Action selection successful. Action keys: {list(action.keys())}")
        
        # Test learning
        for _ in range(50):
            obs = env.reset()
            for _ in range(10):
                action = agent.act(obs, explore=True)
                next_obs, reward, done, info = env.step(action)
                agent.remember(obs, action, reward, next_obs, done)
                obs = next_obs
                if done:
                    break
        
        # Try to learn
        if len(agent.memory) >= agent.config.batch_size:
            loss = agent.learn()
            if loss is not None:
                print(f"Learning successful. Loss: {loss:.4f}")
            else:
                print("Learning skipped (not enough experiences)")
        
        print("Agent test: SUCCESS")
    except Exception as e:
        print(f"Agent test: FAILED - {e}")
        import traceback
        traceback.print_exc()
        return False
    
    print()
    return True


def test_training_script():
    """Test if training script can be imported"""
    print("=== Training Script Test ===")
    
    try:
        from ml.training.train_rl import RLTrainer, load_config
        
        # Check if config exists
        config_path = Path("configs/rl_training.yaml")
        if config_path.exists():
            config = load_config(str(config_path))
            print(f"Config loaded successfully. Episodes: {config['num_episodes']}")
            
            # Create trainer (don't run)
            trainer = RLTrainer(config)
            print("Trainer created successfully")
        else:
            print(f"Config file not found at {config_path}")
            
        print("Training script test: SUCCESS")
    except Exception as e:
        print(f"Training script test: FAILED - {e}")
        return False
    
    print()
    return True


def test_dependencies():
    """Test if all required packages are installed"""
    print("=== Dependencies Test ===")
    
    required_packages = [
        ('torch', 'PyTorch'),
        ('numpy', 'NumPy'),
        ('gym', 'OpenAI Gym'),
        ('mlflow', 'MLflow'),
        ('tqdm', 'tqdm'),
        ('yaml', 'PyYAML')
    ]
    
    all_installed = True
    for package, name in required_packages:
        try:
            __import__(package)
            print(f"✓ {name} installed")
        except ImportError:
            print(f"✗ {name} NOT installed")
            all_installed = False
    
    print(f"\nDependencies test: {'SUCCESS' if all_installed else 'FAILED'}")
    print()
    return all_installed


def main():
    """Run all tests"""
    print("LLMrecon ML Setup Test")
    print("=" * 50)
    print()
    
    results = {
        'Dependencies': test_dependencies(),
        'GPU': test_gpu(),
        'Environment': test_environment(),
        'Agent': test_agent(),
        'Training Script': test_training_script()
    }
    
    print("=" * 50)
    print("Test Summary:")
    for test, passed in results.items():
        status = "PASSED" if passed else "FAILED"
        print(f"  {test}: {status}")
    
    all_passed = all(results.values())
    print()
    if all_passed:
        print("✅ All tests passed! ML environment is ready.")
        print("\nNext steps:")
        print("1. Run training: python -m ml.training.train_rl")
        print("2. Monitor with MLflow: mlflow ui")
        print("3. View tensorboard: tensorboard --logdir logs")
    else:
        print("❌ Some tests failed. Please fix the issues above.")
        print("\nCommon fixes:")
        print("- Install missing packages: pip install -r requirements-ml.txt")
        print("- Check CUDA installation: nvidia-smi")
        print("- Verify Python path: echo $PYTHONPATH")


if __name__ == '__main__':
    main()