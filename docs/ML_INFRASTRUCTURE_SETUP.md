# ML Infrastructure Setup Guide for v0.3.0

## 🚀 Quick Start

```bash
# Clone and setup
cd llmrecon
make ml-setup

# Verify GPU
nvidia-smi

# Test ML environment
python -m llmrecon.ml.test_setup
```

## 📋 Prerequisites

### Hardware Requirements
- **Development**: 1x NVIDIA GPU (8GB+ VRAM)
- **Training**: 2-4x V100/A100 GPUs
- **Inference**: 1x T4/V100 GPU
- **CPU**: 16+ cores, 32GB+ RAM

### Software Requirements
```yaml
system:
  - Ubuntu 20.04+ or macOS 12+
  - CUDA 11.8+ (for GPU)
  - Docker 20.10+
  - Kubernetes 1.25+

python:
  - Python 3.9+
  - pip 22.0+
  - virtualenv/conda
```

## 🔧 Installation Steps

### 1. GPU Setup (Linux)

```bash
# Install NVIDIA drivers
sudo apt update
sudo apt install nvidia-driver-525

# Install CUDA Toolkit
wget https://developer.download.nvidia.com/compute/cuda/11.8.0/local_installers/cuda_11.8.0_520.61.05_linux.run
sudo sh cuda_11.8.0_520.61.05_linux.run

# Install cuDNN
# Download from https://developer.nvidia.com/cudnn
sudo dpkg -i cudnn-local-repo-ubuntu2004-8.9.0.131_1.0-1_amd64.deb

# Verify installation
nvidia-smi
nvcc --version
```

### 2. Python Environment

```bash
# Create virtual environment
python3 -m venv venv-ml
source venv-ml/bin/activate

# Upgrade pip
pip install --upgrade pip setuptools wheel

# Install ML frameworks
pip install torch==2.0.1+cu118 torchvision==0.15.2+cu118 -f https://download.pytorch.org/whl/torch_stable.html
pip install tensorflow==2.12.0
pip install transformers==4.30.0
pip install gym==0.26.2
pip install stable-baselines3==2.0.0
```

### 3. Development Tools

```bash
# Jupyter for experimentation
pip install jupyter jupyterlab ipywidgets

# Experiment tracking
pip install mlflow==2.4.0 wandb==0.15.0

# Data processing
pip install pandas numpy scikit-learn matplotlib seaborn

# Distributed training
pip install ray[default]==2.5.0 horovod==0.28.0

# Model serving
pip install torchserve torch-model-archiver
```

### 4. Infrastructure Components

#### Redis for Caching
```bash
# Using Docker
docker run -d --name redis-ml \
  -p 6379:6379 \
  redis:7-alpine \
  redis-server --appendonly yes
```

#### MLflow Tracking Server
```bash
# Create MLflow backend store
mkdir -p ~/mlflow

# Start MLflow server
mlflow server \
  --backend-store-uri sqlite:///~/mlflow/mlflow.db \
  --default-artifact-root ~/mlflow/artifacts \
  --host 0.0.0.0 \
  --port 5000
```

#### Feature Store (Feast)
```bash
pip install feast

# Initialize feature store
feast init llm_features
cd llm_features

# Configure feature_store.yaml
cat > feature_store.yaml << EOF
project: llmrecon
registry: data/registry.db
provider: local
online_store:
  type: redis
  connection_string: localhost:6379
EOF
```

## 🏗️ Project Structure

```
llmrecon/
├── ml/
│   ├── __init__.py
│   ├── environments/          # RL environments
│   │   ├── attack_env.py
│   │   └── llm_gym.py
│   ├── agents/               # RL agents
│   │   ├── dqn.py
│   │   ├── multi_armed_bandit.py
│   │   └── base_agent.py
│   ├── models/               # Neural networks
│   │   ├── attack_generator.py
│   │   ├── discriminator.py
│   │   └── embeddings.py
│   ├── training/             # Training scripts
│   │   ├── train_rl.py
│   │   ├── train_generator.py
│   │   └── distributed.py
│   ├── inference/            # Inference services
│   │   ├── model_server.py
│   │   └── cache.py
│   ├── data/                 # Data processing
│   │   ├── features.py
│   │   ├── preprocessing.py
│   │   └── augmentation.py
│   └── utils/                # Utilities
│       ├── metrics.py
│       ├── visualization.py
│       └── config.py
├── notebooks/                # Jupyter notebooks
│   ├── 01_rl_exploration.ipynb
│   ├── 02_generator_training.ipynb
│   └── 03_discovery_analysis.ipynb
├── configs/                  # Configuration files
│   ├── ml_config.yaml
│   ├── training_config.yaml
│   └── serving_config.yaml
└── tests/                    # ML tests
    ├── test_environments.py
    ├── test_agents.py
    └── test_models.py
```

## 🎮 Environment Configuration

### RL Environment Setup

```python
# ml/environments/attack_env.py
import gym
import numpy as np
from gym import spaces

class LLMAttackEnv(gym.Env):
    """OpenAI Gym environment for LLM attacks"""
    
    def __init__(self, config):
        super().__init__()
        
        # Define action space (attack parameters)
        self.action_space = spaces.Dict({
            'attack_type': spaces.Discrete(5),
            'intensity': spaces.Box(0, 1, shape=(1,)),
            'target_model': spaces.Discrete(3)
        })
        
        # Define observation space (state)
        self.observation_space = spaces.Dict({
            'model_state': spaces.Box(-1, 1, shape=(128,)),
            'history': spaces.Box(0, 1, shape=(10, 64)),
            'context': spaces.Box(-1, 1, shape=(256,))
        })
        
    def reset(self):
        """Reset environment to initial state"""
        return self.observation_space.sample()
    
    def step(self, action):
        """Execute action and return new state"""
        # Implement attack execution logic
        obs = self.observation_space.sample()
        reward = np.random.random()
        done = False
        info = {}
        return obs, reward, done, info
```

### Training Configuration

```yaml
# configs/training_config.yaml
rl_training:
  algorithm: "DQN"
  episodes: 1000
  batch_size: 32
  learning_rate: 0.001
  gamma: 0.99
  epsilon_start: 1.0
  epsilon_end: 0.01
  epsilon_decay: 0.995
  
  experience_replay:
    capacity: 10000
    prioritized: true
    alpha: 0.6
    beta: 0.4
    
  distributed:
    enabled: true
    num_workers: 4
    gpus_per_worker: 1

generator_training:
  model: "gpt2"
  epochs: 10
  batch_size: 16
  learning_rate: 5e-5
  max_length: 512
  temperature: 0.8
  
  data:
    train_split: 0.8
    validation_split: 0.1
    test_split: 0.1
```

## 🐳 Docker Setup

### ML Development Container

```dockerfile
# Dockerfile.ml
FROM nvidia/cuda:11.8.0-cudnn8-devel-ubuntu20.04

# Install Python and dependencies
RUN apt-get update && apt-get install -y \
    python3.9 python3-pip git wget curl \
    && rm -rf /var/lib/apt/lists/*

# Install ML frameworks
COPY requirements-ml.txt /tmp/
RUN pip3 install -r /tmp/requirements-ml.txt

# Set up workspace
WORKDIR /workspace
COPY . .

# Expose ports
EXPOSE 8888 5000 6006

CMD ["jupyter", "lab", "--ip=0.0.0.0", "--allow-root"]
```

### Docker Compose

```yaml
# docker-compose.ml.yml
version: '3.8'

services:
  ml-dev:
    build:
      context: .
      dockerfile: Dockerfile.ml
    runtime: nvidia
    environment:
      - NVIDIA_VISIBLE_DEVICES=all
    volumes:
      - .:/workspace
      - ~/.cache:/root/.cache
    ports:
      - "8888:8888"  # Jupyter
      - "5000:5000"  # MLflow
      - "6006:6006"  # TensorBoard
    
  redis-ml:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-ml-data:/data
      
  mlflow:
    image: python:3.9
    command: >
      bash -c "pip install mlflow &&
               mlflow server --host 0.0.0.0"
    ports:
      - "5001:5000"
    volumes:
      - mlflow-data:/mlflow

volumes:
  redis-ml-data:
  mlflow-data:
```

## 🧪 Testing the Setup

### 1. GPU Test

```python
# test_gpu.py
import torch
import tensorflow as tf

print(f"PyTorch CUDA available: {torch.cuda.is_available()}")
print(f"PyTorch CUDA devices: {torch.cuda.device_count()}")
print(f"TensorFlow GPUs: {len(tf.config.list_physical_devices('GPU'))}")

# Simple GPU computation test
if torch.cuda.is_available():
    x = torch.randn(1000, 1000).cuda()
    y = torch.randn(1000, 1000).cuda()
    z = torch.matmul(x, y)
    print(f"GPU computation successful: {z.shape}")
```

### 2. RL Environment Test

```python
# test_rl_env.py
from ml.environments.attack_env import LLMAttackEnv

env = LLMAttackEnv({})
obs = env.reset()
print(f"Initial observation: {obs}")

for i in range(10):
    action = env.action_space.sample()
    obs, reward, done, info = env.step(action)
    print(f"Step {i}: reward={reward:.3f}")
```

### 3. Model Serving Test

```bash
# Start TorchServe
torchserve --start --model-store model_store --models all

# Test inference
curl -X POST http://localhost:8080/predictions/attack_generator \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Generate an attack payload"}'
```

## 📚 Additional Resources

### Tutorials
- [RL with Stable Baselines3](https://stable-baselines3.readthedocs.io/)
- [Transformers Fine-tuning](https://huggingface.co/docs/transformers/training)
- [MLflow Tracking](https://mlflow.org/docs/latest/tracking.html)

### Best Practices
- Always use GPU for training, CPU for development
- Version control models with DVC or MLflow
- Monitor GPU memory usage during training
- Use mixed precision training for efficiency

## 🎉 Next Steps

1. Run `make ml-setup` to install everything
2. Start Jupyter Lab and explore notebooks
3. Begin implementing the RL environment
4. Set up MLflow for experiment tracking

---

*With this infrastructure in place, we're ready to build the AI-powered features for v0.3.0!*