"""
GAN-style Discriminator for Attack Generation in LLMrecon

This module implements a Generative Adversarial Network (GAN) approach
for generating and discriminating attack payloads, creating more sophisticated
and hard-to-detect attacks.
"""

import torch
import torch.nn as nn
import torch.nn.functional as F
import torch.optim as optim
from torch.utils.data import DataLoader, TensorDataset
import numpy as np
from typing import List, Dict, Any, Tuple, Optional
from dataclasses import dataclass
import json
import logging
from collections import deque
import matplotlib.pyplot as plt

logger = logging.getLogger(__name__)


@dataclass
class GANConfig:
    """Configuration for GAN models"""
    # Model dimensions
    vocab_size: int = 10000
    embedding_dim: int = 128
    hidden_dim: int = 256
    latent_dim: int = 100
    max_seq_length: int = 100
    
    # Training parameters
    learning_rate_g: float = 0.0002
    learning_rate_d: float = 0.0002
    beta1: float = 0.5
    beta2: float = 0.999
    
    # Architecture choices
    num_layers: int = 3
    dropout: float = 0.2
    use_attention: bool = True
    
    # Training strategy
    d_steps_per_g_step: int = 2
    label_smoothing: float = 0.1
    gradient_penalty_weight: float = 10.0


class SelfAttention(nn.Module):
    """Self-attention layer for sequence modeling"""
    
    def __init__(self, hidden_dim: int):
        super().__init__()
        self.hidden_dim = hidden_dim
        self.query = nn.Linear(hidden_dim, hidden_dim)
        self.key = nn.Linear(hidden_dim, hidden_dim)
        self.value = nn.Linear(hidden_dim, hidden_dim)
        self.scale = np.sqrt(hidden_dim)
        
    def forward(self, x: torch.Tensor, mask: Optional[torch.Tensor] = None) -> torch.Tensor:
        batch_size, seq_len, _ = x.shape
        
        # Compute Q, K, V
        q = self.query(x)
        k = self.key(x)
        v = self.value(x)
        
        # Compute attention scores
        scores = torch.bmm(q, k.transpose(1, 2)) / self.scale
        
        # Apply mask if provided
        if mask is not None:
            scores = scores.masked_fill(mask == 0, -1e9)
        
        # Apply softmax
        attn_weights = F.softmax(scores, dim=-1)
        
        # Apply attention to values
        output = torch.bmm(attn_weights, v)
        
        return output


class AttackGenerator(nn.Module):
    """
    Generator network for creating attack payloads.
    
    Features:
    - LSTM-based sequence generation
    - Conditional generation based on attack type
    - Self-attention for long-range dependencies
    """
    
    def __init__(self, config: GANConfig):
        super().__init__()
        self.config = config
        
        # Noise projection
        self.noise_projection = nn.Sequential(
            nn.Linear(config.latent_dim, config.hidden_dim),
            nn.ReLU(),
            nn.Dropout(config.dropout)
        )
        
        # Embeddings
        self.token_embedding = nn.Embedding(config.vocab_size, config.embedding_dim)
        self.attack_type_embedding = nn.Embedding(10, config.embedding_dim)  # 10 attack types
        
        # LSTM layers
        self.lstm = nn.LSTM(
            config.embedding_dim + config.hidden_dim,
            config.hidden_dim,
            config.num_layers,
            batch_first=True,
            dropout=config.dropout if config.num_layers > 1 else 0
        )
        
        # Self-attention
        if config.use_attention:
            self.attention = SelfAttention(config.hidden_dim)
        
        # Output projection
        self.output_projection = nn.Linear(config.hidden_dim, config.vocab_size)
        
        # Layer normalization
        self.layer_norm = nn.LayerNorm(config.hidden_dim)
        
    def forward(self, 
                noise: torch.Tensor,
                attack_type: torch.Tensor,
                temperature: float = 1.0) -> torch.Tensor:
        """
        Generate attack sequences from noise.
        
        Args:
            noise: Random noise [batch_size, latent_dim]
            attack_type: Attack type IDs [batch_size]
            temperature: Sampling temperature
            
        Returns:
            Generated sequences [batch_size, seq_length, vocab_size]
        """
        batch_size = noise.shape[0]
        device = noise.device
        
        # Project noise
        h = self.noise_projection(noise)  # [batch_size, hidden_dim]
        
        # Get attack type embedding
        attack_embed = self.attack_type_embedding(attack_type)  # [batch_size, embedding_dim]
        
        # Initialize hidden state with noise
        h_0 = h.unsqueeze(0).repeat(self.config.num_layers, 1, 1)
        c_0 = torch.zeros_like(h_0)
        
        # Start with special start token (assuming 1 is START)
        input_seq = torch.ones(batch_size, 1, dtype=torch.long, device=device)
        
        outputs = []
        hidden = (h_0, c_0)
        
        for _ in range(self.config.max_seq_length):
            # Get token embeddings
            token_embed = self.token_embedding(input_seq[:, -1:])  # [batch_size, 1, embedding_dim]
            
            # Combine with attack embedding and noise
            combined_input = torch.cat([
                token_embed + attack_embed.unsqueeze(1),
                h.unsqueeze(1)
            ], dim=-1)
            
            # LSTM forward
            lstm_out, hidden = self.lstm(combined_input, hidden)
            
            # Apply attention if enabled
            if self.config.use_attention and len(outputs) > 0:
                # Concatenate all previous outputs for attention
                all_outputs = torch.cat(outputs, dim=1)
                attn_out = self.attention(all_outputs)
                lstm_out = lstm_out + attn_out[:, -1:, :]
            
            # Layer norm
            lstm_out = self.layer_norm(lstm_out)
            
            # Project to vocabulary
            logits = self.output_projection(lstm_out) / temperature
            
            # Sample next token
            probs = F.softmax(logits, dim=-1)
            next_token = torch.multinomial(probs.squeeze(1), 1)
            
            # Append to sequence
            input_seq = torch.cat([input_seq, next_token], dim=1)
            outputs.append(lstm_out)
            
            # Stop if all sequences have generated END token (assuming 2 is END)
            if (next_token == 2).all():
                break
        
        # Stack all outputs
        all_outputs = torch.cat(outputs, dim=1)
        final_logits = self.output_projection(all_outputs)
        
        return final_logits


class AttackDiscriminator(nn.Module):
    """
    Discriminator network for distinguishing real vs generated attacks.
    
    Features:
    - CNN + LSTM architecture
    - Multi-scale feature extraction
    - Adversarial robustness features
    """
    
    def __init__(self, config: GANConfig):
        super().__init__()
        self.config = config
        
        # Token embedding
        self.embedding = nn.Embedding(config.vocab_size, config.embedding_dim)
        
        # CNN layers for local pattern detection
        self.conv_layers = nn.ModuleList([
            nn.Conv1d(config.embedding_dim, config.hidden_dim // 2, kernel_size=3, padding=1),
            nn.Conv1d(config.embedding_dim, config.hidden_dim // 2, kernel_size=5, padding=2),
            nn.Conv1d(config.embedding_dim, config.hidden_dim // 2, kernel_size=7, padding=3)
        ])
        
        # LSTM for sequence modeling
        self.lstm = nn.LSTM(
            config.hidden_dim * 3 // 2,  # Concatenated conv outputs
            config.hidden_dim,
            config.num_layers,
            batch_first=True,
            bidirectional=True,
            dropout=config.dropout if config.num_layers > 1 else 0
        )
        
        # Attention pooling
        if config.use_attention:
            self.attention = SelfAttention(config.hidden_dim * 2)  # Bidirectional
        
        # Classification layers
        self.classifier = nn.Sequential(
            nn.Linear(config.hidden_dim * 2, config.hidden_dim),
            nn.ReLU(),
            nn.Dropout(config.dropout),
            nn.Linear(config.hidden_dim, config.hidden_dim // 2),
            nn.ReLU(),
            nn.Dropout(config.dropout),
            nn.Linear(config.hidden_dim // 2, 1)
        )
        
        # Feature extraction layers (for analysis)
        self.feature_extractor = nn.Linear(config.hidden_dim * 2, 64)
        
    def forward(self, 
                sequences: torch.Tensor,
                mask: Optional[torch.Tensor] = None,
                return_features: bool = False) -> Tuple[torch.Tensor, Optional[torch.Tensor]]:
        """
        Discriminate between real and generated sequences.
        
        Args:
            sequences: Token sequences [batch_size, seq_length]
            mask: Attention mask [batch_size, seq_length]
            return_features: Whether to return intermediate features
            
        Returns:
            predictions: Realness scores [batch_size, 1]
            features: Optional intermediate features [batch_size, 64]
        """
        # Embed tokens
        embedded = self.embedding(sequences)  # [batch_size, seq_length, embedding_dim]
        
        # Apply CNN layers
        # Transpose for conv1d: [batch_size, embedding_dim, seq_length]
        embedded_t = embedded.transpose(1, 2)
        
        conv_outputs = []
        for conv in self.conv_layers:
            conv_out = F.relu(conv(embedded_t))
            conv_outputs.append(conv_out)
        
        # Concatenate conv outputs and transpose back
        conv_combined = torch.cat(conv_outputs, dim=1)  # [batch_size, hidden_dim * 3/2, seq_length]
        conv_combined = conv_combined.transpose(1, 2)  # [batch_size, seq_length, hidden_dim * 3/2]
        
        # LSTM processing
        lstm_out, _ = self.lstm(conv_combined)  # [batch_size, seq_length, hidden_dim * 2]
        
        # Apply attention if enabled
        if self.config.use_attention:
            lstm_out = self.attention(lstm_out, mask)
        
        # Global pooling (mean and max)
        if mask is not None:
            # Masked mean pooling
            mask_expanded = mask.unsqueeze(-1).expand_as(lstm_out)
            masked_output = lstm_out * mask_expanded
            sum_output = masked_output.sum(dim=1)
            lengths = mask.sum(dim=1, keepdim=True).float()
            mean_pooled = sum_output / lengths
        else:
            mean_pooled = lstm_out.mean(dim=1)
        
        max_pooled, _ = lstm_out.max(dim=1)
        
        # Combine pooling strategies
        pooled = (mean_pooled + max_pooled) / 2
        
        # Extract features if requested
        features = None
        if return_features:
            features = self.feature_extractor(pooled)
        
        # Classification
        predictions = self.classifier(pooled)
        
        return predictions, features


class AttackGAN:
    """
    Main GAN system for attack generation and discrimination.
    
    Features:
    - Wasserstein GAN with gradient penalty (WGAN-GP)
    - Conditional generation based on attack types
    - Progressive training strategies
    - Attack quality metrics
    """
    
    def __init__(self, config: Optional[GANConfig] = None):
        self.config = config or GANConfig()
        self.device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
        
        # Initialize networks
        self.generator = AttackGenerator(self.config).to(self.device)
        self.discriminator = AttackDiscriminator(self.config).to(self.device)
        
        # Optimizers
        self.optimizer_g = optim.Adam(
            self.generator.parameters(),
            lr=self.config.learning_rate_g,
            betas=(self.config.beta1, self.config.beta2)
        )
        
        self.optimizer_d = optim.Adam(
            self.discriminator.parameters(),
            lr=self.config.learning_rate_d,
            betas=(self.config.beta1, self.config.beta2)
        )
        
        # Training history
        self.g_losses = deque(maxlen=1000)
        self.d_losses = deque(maxlen=1000)
        self.d_real_scores = deque(maxlen=1000)
        self.d_fake_scores = deque(maxlen=1000)
        
        # Best generated attacks
        self.best_attacks = []
        
    def _compute_gradient_penalty(self, real_samples: torch.Tensor, fake_samples: torch.Tensor) -> torch.Tensor:
        """Compute gradient penalty for WGAN-GP"""
        batch_size = real_samples.size(0)
        
        # Random weight for interpolation
        alpha = torch.rand(batch_size, 1, device=self.device)
        alpha = alpha.expand_as(real_samples)
        
        # Interpolate between real and fake samples
        interpolated = alpha * real_samples + (1 - alpha) * fake_samples
        interpolated.requires_grad_(True)
        
        # Get discriminator output for interpolated samples
        d_interpolated, _ = self.discriminator(interpolated)
        
        # Compute gradients
        gradients = torch.autograd.grad(
            outputs=d_interpolated,
            inputs=interpolated,
            grad_outputs=torch.ones_like(d_interpolated),
            create_graph=True,
            retain_graph=True,
            only_inputs=True
        )[0]
        
        # Compute gradient penalty
        gradients = gradients.view(batch_size, -1)
        gradient_penalty = ((gradients.norm(2, dim=1) - 1) ** 2).mean()
        
        return gradient_penalty
    
    def train_step(self, 
                   real_sequences: torch.Tensor,
                   attack_types: torch.Tensor,
                   masks: Optional[torch.Tensor] = None) -> Dict[str, float]:
        """
        Single training step for GAN.
        
        Args:
            real_sequences: Real attack sequences [batch_size, seq_length]
            attack_types: Attack type IDs [batch_size]
            masks: Attention masks [batch_size, seq_length]
            
        Returns:
            Dictionary of losses and scores
        """
        batch_size = real_sequences.size(0)
        
        # Train discriminator
        for _ in range(self.config.d_steps_per_g_step):
            self.optimizer_d.zero_grad()
            
            # Real samples
            d_real, _ = self.discriminator(real_sequences, masks)
            
            # Generate fake samples
            noise = torch.randn(batch_size, self.config.latent_dim, device=self.device)
            with torch.no_grad():
                fake_logits = self.generator(noise, attack_types)
                fake_sequences = fake_logits.argmax(dim=-1)
            
            # Fake samples
            d_fake, _ = self.discriminator(fake_sequences)
            
            # Wasserstein loss
            d_loss = -torch.mean(d_real) + torch.mean(d_fake)
            
            # Gradient penalty
            gp = self._compute_gradient_penalty(real_sequences.float(), fake_sequences.float())
            d_loss += self.config.gradient_penalty_weight * gp
            
            d_loss.backward()
            self.optimizer_d.step()
        
        # Train generator
        self.optimizer_g.zero_grad()
        
        noise = torch.randn(batch_size, self.config.latent_dim, device=self.device)
        fake_logits = self.generator(noise, attack_types)
        fake_sequences = fake_logits.argmax(dim=-1)
        
        d_fake_for_g, features = self.discriminator(fake_sequences, return_features=True)
        
        # Generator loss (wants to maximize discriminator output)
        g_loss = -torch.mean(d_fake_for_g)
        
        # Feature matching loss (optional)
        if features is not None:
            _, real_features = self.discriminator(real_sequences, masks, return_features=True)
            feature_loss = F.mse_loss(features, real_features.detach())
            g_loss += 0.1 * feature_loss
        
        g_loss.backward()
        self.optimizer_g.step()
        
        # Record metrics
        self.g_losses.append(g_loss.item())
        self.d_losses.append(d_loss.item())
        self.d_real_scores.append(d_real.mean().item())
        self.d_fake_scores.append(d_fake.mean().item())
        
        return {
            'g_loss': g_loss.item(),
            'd_loss': d_loss.item(),
            'd_real_score': d_real.mean().item(),
            'd_fake_score': d_fake.mean().item(),
            'gradient_penalty': gp.item()
        }
    
    def generate_attacks(self,
                        num_samples: int,
                        attack_type: int,
                        temperature: float = 0.8) -> List[torch.Tensor]:
        """Generate attack sequences"""
        self.generator.eval()
        
        with torch.no_grad():
            noise = torch.randn(num_samples, self.config.latent_dim, device=self.device)
            attack_types = torch.full((num_samples,), attack_type, dtype=torch.long, device=self.device)
            
            fake_logits = self.generator(noise, attack_types, temperature)
            fake_sequences = fake_logits.argmax(dim=-1)
        
        self.generator.train()
        return fake_sequences.cpu().tolist()
    
    def evaluate_quality(self, generated_sequences: List[torch.Tensor]) -> Dict[str, float]:
        """Evaluate quality of generated attacks"""
        self.discriminator.eval()
        
        quality_scores = []
        diversity_scores = []
        
        with torch.no_grad():
            for seq in generated_sequences:
                seq_tensor = torch.tensor(seq, device=self.device).unsqueeze(0)
                score, features = self.discriminator(seq_tensor, return_features=True)
                quality_scores.append(score.item())
                
                if features is not None:
                    diversity_scores.append(features.cpu().numpy())
        
        self.discriminator.train()
        
        # Calculate diversity as average pairwise distance
        if diversity_scores:
            diversity_scores = np.array(diversity_scores)
            pairwise_distances = []
            for i in range(len(diversity_scores)):
                for j in range(i + 1, len(diversity_scores)):
                    dist = np.linalg.norm(diversity_scores[i] - diversity_scores[j])
                    pairwise_distances.append(dist)
            
            avg_diversity = np.mean(pairwise_distances) if pairwise_distances else 0
        else:
            avg_diversity = 0
        
        return {
            'avg_quality_score': np.mean(quality_scores),
            'std_quality_score': np.std(quality_scores),
            'avg_diversity': avg_diversity,
            'num_unique': len(set(str(seq) for seq in generated_sequences))
        }
    
    def save_checkpoint(self, path: str):
        """Save model checkpoint"""
        torch.save({
            'generator_state': self.generator.state_dict(),
            'discriminator_state': self.discriminator.state_dict(),
            'optimizer_g_state': self.optimizer_g.state_dict(),
            'optimizer_d_state': self.optimizer_d.state_dict(),
            'config': self.config,
            'g_losses': list(self.g_losses),
            'd_losses': list(self.d_losses)
        }, path)
        logger.info(f"Checkpoint saved to {path}")
    
    def load_checkpoint(self, path: str):
        """Load model checkpoint"""
        checkpoint = torch.load(path, map_location=self.device)
        
        self.generator.load_state_dict(checkpoint['generator_state'])
        self.discriminator.load_state_dict(checkpoint['discriminator_state'])
        self.optimizer_g.load_state_dict(checkpoint['optimizer_g_state'])
        self.optimizer_d.load_state_dict(checkpoint['optimizer_d_state'])
        
        self.g_losses = deque(checkpoint['g_losses'], maxlen=1000)
        self.d_losses = deque(checkpoint['d_losses'], maxlen=1000)
        
        logger.info(f"Checkpoint loaded from {path}")
    
    def plot_training_history(self) -> plt.Figure:
        """Plot training history"""
        fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(10, 8))
        
        # Losses
        ax1.plot(self.g_losses, label='Generator Loss', alpha=0.7)
        ax1.plot(self.d_losses, label='Discriminator Loss', alpha=0.7)
        ax1.set_xlabel('Training Step')
        ax1.set_ylabel('Loss')
        ax1.set_title('GAN Training Losses')
        ax1.legend()
        ax1.grid(True, alpha=0.3)
        
        # Scores
        ax2.plot(self.d_real_scores, label='Real Score', alpha=0.7)
        ax2.plot(self.d_fake_scores, label='Fake Score', alpha=0.7)
        ax2.set_xlabel('Training Step')
        ax2.set_ylabel('Discriminator Score')
        ax2.set_title('Discriminator Scores')
        ax2.legend()
        ax2.grid(True, alpha=0.3)
        
        plt.tight_layout()
        return fig


def create_attack_gan(config: Optional[GANConfig] = None) -> AttackGAN:
    """Create and initialize AttackGAN"""
    return AttackGAN(config)