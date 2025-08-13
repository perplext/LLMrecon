"""
Transformer-based Attack Generator for LLMrecon

This module implements a transformer-based model for generating sophisticated
attack payloads using attention mechanisms and learned patterns.
"""

import torch
import torch.nn as nn
import torch.nn.functional as F
from torch.nn import TransformerEncoder, TransformerEncoderLayer
import numpy as np
from typing import List, Dict, Any, Optional, Tuple
from dataclasses import dataclass
import json
import math
from collections import Counter
import logging

logger = logging.getLogger(__name__)


@dataclass
class TransformerConfig:
    """Configuration for transformer model"""
    vocab_size: int = 10000
    d_model: int = 512
    nhead: int = 8
    num_layers: int = 6
    dim_feedforward: int = 2048
    dropout: float = 0.1
    max_seq_length: int = 512
    num_attack_types: int = 10
    num_target_models: int = 5


class PositionalEncoding(nn.Module):
    """Add positional encoding to embeddings"""
    
    def __init__(self, d_model: int, dropout: float = 0.1, max_len: int = 5000):
        super().__init__()
        self.dropout = nn.Dropout(p=dropout)
        
        pe = torch.zeros(max_len, d_model)
        position = torch.arange(0, max_len, dtype=torch.float).unsqueeze(1)
        div_term = torch.exp(torch.arange(0, d_model, 2).float() * 
                           (-math.log(10000.0) / d_model))
        
        pe[:, 0::2] = torch.sin(position * div_term)
        pe[:, 1::2] = torch.cos(position * div_term)
        pe = pe.unsqueeze(0).transpose(0, 1)
        
        self.register_buffer('pe', pe)
    
    def forward(self, x: torch.Tensor) -> torch.Tensor:
        x = x + self.pe[:x.size(0), :]
        return self.dropout(x)


class AttackTransformer(nn.Module):
    """
    Transformer model for attack generation.
    
    Features:
    - Multi-head self-attention
    - Conditional generation based on attack type and target
    - Learned embeddings for attack patterns
    - Controllable generation parameters
    """
    
    def __init__(self, config: TransformerConfig):
        super().__init__()
        self.config = config
        
        # Token embeddings
        self.token_embedding = nn.Embedding(config.vocab_size, config.d_model)
        self.pos_encoder = PositionalEncoding(config.d_model, config.dropout)
        
        # Attack type and target embeddings
        self.attack_type_embedding = nn.Embedding(config.num_attack_types, config.d_model)
        self.target_embedding = nn.Embedding(config.num_target_models, config.d_model)
        
        # Transformer encoder
        encoder_layers = TransformerEncoderLayer(
            config.d_model,
            config.nhead,
            config.dim_feedforward,
            config.dropout,
            batch_first=True
        )
        self.transformer = TransformerEncoder(encoder_layers, config.num_layers)
        
        # Output projection
        self.output_projection = nn.Linear(config.d_model, config.vocab_size)
        
        # Generation control parameters
        self.style_embedding = nn.Linear(4, config.d_model)  # 4 style parameters
        
        self._init_weights()
    
    def _init_weights(self):
        """Initialize weights"""
        init_range = 0.1
        self.token_embedding.weight.data.uniform_(-init_range, init_range)
        self.attack_type_embedding.weight.data.uniform_(-init_range, init_range)
        self.target_embedding.weight.data.uniform_(-init_range, init_range)
        self.output_projection.bias.data.zero_()
        self.output_projection.weight.data.uniform_(-init_range, init_range)
    
    def forward(self, 
                input_ids: torch.Tensor,
                attack_type: torch.Tensor,
                target_model: torch.Tensor,
                style_params: Optional[torch.Tensor] = None,
                attention_mask: Optional[torch.Tensor] = None) -> torch.Tensor:
        """
        Forward pass through the transformer.
        
        Args:
            input_ids: Token IDs [batch_size, seq_len]
            attack_type: Attack type IDs [batch_size]
            target_model: Target model IDs [batch_size]
            style_params: Style control parameters [batch_size, 4]
            attention_mask: Attention mask [batch_size, seq_len]
            
        Returns:
            Logits over vocabulary [batch_size, seq_len, vocab_size]
        """
        batch_size, seq_len = input_ids.shape
        
        # Get embeddings
        token_embeds = self.token_embedding(input_ids) * math.sqrt(self.config.d_model)
        token_embeds = self.pos_encoder(token_embeds.transpose(0, 1)).transpose(0, 1)
        
        # Add attack type and target embeddings
        attack_embeds = self.attack_type_embedding(attack_type).unsqueeze(1)
        target_embeds = self.target_embedding(target_model).unsqueeze(1)
        
        # Combine embeddings
        combined_embeds = token_embeds + attack_embeds + target_embeds
        
        # Add style embeddings if provided
        if style_params is not None:
            style_embeds = self.style_embedding(style_params).unsqueeze(1)
            combined_embeds = combined_embeds + style_embeds
        
        # Apply transformer
        if attention_mask is not None:
            # Convert padding mask to attention mask
            attention_mask = attention_mask.float()
            attention_mask = attention_mask.masked_fill(attention_mask == 0, float('-inf'))
            attention_mask = attention_mask.masked_fill(attention_mask == 1, float(0.0))
        
        transformer_output = self.transformer(combined_embeds, src_key_padding_mask=attention_mask)
        
        # Project to vocabulary
        output = self.output_projection(transformer_output)
        
        return output
    
    def generate(self,
                 prompt: torch.Tensor,
                 attack_type: torch.Tensor,
                 target_model: torch.Tensor,
                 max_length: int = 100,
                 temperature: float = 1.0,
                 top_k: int = 50,
                 top_p: float = 0.95,
                 style_params: Optional[torch.Tensor] = None) -> torch.Tensor:
        """
        Generate attack payload using the transformer.
        
        Args:
            prompt: Starting tokens [batch_size, prompt_len]
            attack_type: Attack type ID [batch_size]
            target_model: Target model ID [batch_size]
            max_length: Maximum generation length
            temperature: Sampling temperature
            top_k: Top-k sampling parameter
            top_p: Top-p (nucleus) sampling parameter
            style_params: Style control parameters
            
        Returns:
            Generated token IDs [batch_size, seq_len]
        """
        self.eval()
        batch_size = prompt.shape[0]
        device = prompt.device
        
        # Start with prompt
        generated = prompt
        
        with torch.no_grad():
            for _ in range(max_length - prompt.shape[1]):
                # Get model predictions
                outputs = self.forward(
                    generated,
                    attack_type,
                    target_model,
                    style_params
                )
                
                # Get next token logits
                next_token_logits = outputs[:, -1, :] / temperature
                
                # Apply top-k filtering
                if top_k > 0:
                    indices_to_remove = next_token_logits < torch.topk(next_token_logits, top_k)[0][..., -1, None]
                    next_token_logits[indices_to_remove] = float('-inf')
                
                # Apply top-p filtering
                if top_p < 1.0:
                    sorted_logits, sorted_indices = torch.sort(next_token_logits, descending=True)
                    cumulative_probs = torch.cumsum(F.softmax(sorted_logits, dim=-1), dim=-1)
                    
                    # Remove tokens with cumulative probability above threshold
                    sorted_indices_to_remove = cumulative_probs > top_p
                    sorted_indices_to_remove[..., 1:] = sorted_indices_to_remove[..., :-1].clone()
                    sorted_indices_to_remove[..., 0] = 0
                    
                    indices_to_remove = sorted_indices_to_remove.scatter(
                        dim=1, index=sorted_indices, src=sorted_indices_to_remove
                    )
                    next_token_logits[indices_to_remove] = float('-inf')
                
                # Sample next token
                probs = F.softmax(next_token_logits, dim=-1)
                next_token = torch.multinomial(probs, num_samples=1)
                
                # Append to generated sequence
                generated = torch.cat([generated, next_token], dim=1)
                
                # Stop if all sequences have generated end token (assuming 2 is EOS)
                if (next_token == 2).all():
                    break
        
        return generated


class AttackTokenizer:
    """
    Simple tokenizer for attack payloads.
    
    Features:
    - Subword tokenization
    - Special tokens for attack patterns
    - Vocabulary management
    """
    
    def __init__(self, vocab_size: int = 10000):
        self.vocab_size = vocab_size
        self.word_to_idx = {}
        self.idx_to_word = {}
        
        # Special tokens
        self.special_tokens = {
            '<PAD>': 0,
            '<START>': 1,
            '<END>': 2,
            '<UNK>': 3,
            '<MASK>': 4,
            '<INST>': 5,  # Instruction marker
            '<SYS>': 6,   # System prompt marker
            '<OBF>': 7,   # Obfuscation marker
        }
        
        # Initialize with special tokens
        for token, idx in self.special_tokens.items():
            self.word_to_idx[token] = idx
            self.idx_to_word[idx] = token
        
        self.next_idx = len(self.special_tokens)
    
    def build_vocabulary(self, texts: List[str]):
        """Build vocabulary from texts"""
        word_freq = Counter()
        
        # Count word frequencies
        for text in texts:
            words = self._tokenize_text(text)
            word_freq.update(words)
        
        # Add most common words to vocabulary
        most_common = word_freq.most_common(self.vocab_size - len(self.special_tokens))
        
        for word, _ in most_common:
            if word not in self.word_to_idx:
                self.word_to_idx[word] = self.next_idx
                self.idx_to_word[self.next_idx] = word
                self.next_idx += 1
    
    def _tokenize_text(self, text: str) -> List[str]:
        """Simple word-level tokenization"""
        # Handle special patterns
        text = text.replace('[INST]', ' <INST> ')
        text = text.replace('[/INST]', ' ')
        text = text.replace('<<SYS>>', ' <SYS> ')
        text = text.replace('<</SYS>>', ' ')
        
        # Split on whitespace and punctuation
        import re
        words = re.findall(r'\w+|[^\w\s]', text.lower())
        
        return words
    
    def encode(self, text: str, max_length: Optional[int] = None) -> List[int]:
        """Encode text to token IDs"""
        words = self._tokenize_text(text)
        
        # Add start token
        tokens = [self.special_tokens['<START>']]
        
        # Encode words
        for word in words:
            if word in self.word_to_idx:
                tokens.append(self.word_to_idx[word])
            else:
                tokens.append(self.special_tokens['<UNK>'])
        
        # Add end token
        tokens.append(self.special_tokens['<END>'])
        
        # Truncate or pad if needed
        if max_length is not None:
            if len(tokens) > max_length:
                tokens = tokens[:max_length-1] + [self.special_tokens['<END>']]
            else:
                tokens = tokens + [self.special_tokens['<PAD>']] * (max_length - len(tokens))
        
        return tokens
    
    def decode(self, token_ids: List[int], skip_special: bool = True) -> str:
        """Decode token IDs to text"""
        words = []
        
        for token_id in token_ids:
            if token_id in self.idx_to_word:
                word = self.idx_to_word[token_id]
                
                if skip_special and word in self.special_tokens:
                    continue
                
                words.append(word)
            else:
                words.append('<UNK>')
        
        # Join words
        text = ' '.join(words)
        
        # Clean up spacing around punctuation
        text = re.sub(r'\s+([.,!?;:])', r'\1', text)
        text = re.sub(r'([(\[{])\s+', r'\1', text)
        text = re.sub(r'\s+([)\]}])', r'\1', text)
        
        return text


class TransformerAttackGenerator:
    """
    High-level interface for transformer-based attack generation.
    
    Features:
    - Training on attack datasets
    - Controlled generation with multiple parameters
    - Fine-tuning on successful attacks
    - Attack pattern learning
    """
    
    def __init__(self, config: Optional[TransformerConfig] = None):
        self.config = config or TransformerConfig()
        self.model = AttackTransformer(self.config)
        self.tokenizer = AttackTokenizer(self.config.vocab_size)
        self.device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
        self.model.to(self.device)
        
        # Attack type mapping
        self.attack_types = {
            'prompt_injection': 0,
            'jailbreak': 1,
            'data_extraction': 2,
            'role_play': 3,
            'instruction_bypass': 4,
            'context_manipulation': 5,
            'encoding_exploit': 6,
            'multi_turn': 7,
            'indirect_prompt': 8,
            'adversarial': 9
        }
        
        # Target model mapping
        self.target_models = {
            'gpt-3.5-turbo': 0,
            'gpt-4': 1,
            'claude-2': 2,
            'llama-2': 3,
            'gemini-pro': 4
        }
    
    def train_on_dataset(self, 
                        train_data: List[Dict[str, Any]],
                        val_data: Optional[List[Dict[str, Any]]] = None,
                        epochs: int = 10,
                        batch_size: int = 32,
                        learning_rate: float = 1e-4):
        """
        Train the transformer on attack data.
        
        Args:
            train_data: List of dicts with 'payload', 'attack_type', 'target_model', 'success'
            val_data: Optional validation data
            epochs: Number of training epochs
            batch_size: Batch size
            learning_rate: Learning rate
        """
        # Build vocabulary
        all_payloads = [d['payload'] for d in train_data]
        self.tokenizer.build_vocabulary(all_payloads)
        
        # Prepare data
        train_dataset = self._prepare_dataset(train_data)
        val_dataset = self._prepare_dataset(val_data) if val_data else None
        
        # Create data loaders
        train_loader = torch.utils.data.DataLoader(
            train_dataset, batch_size=batch_size, shuffle=True
        )
        
        # Optimizer
        optimizer = torch.optim.Adam(self.model.parameters(), lr=learning_rate)
        criterion = nn.CrossEntropyLoss(ignore_index=self.tokenizer.special_tokens['<PAD>'])
        
        # Training loop
        self.model.train()
        for epoch in range(epochs):
            total_loss = 0
            
            for batch in train_loader:
                input_ids = batch['input_ids'].to(self.device)
                attack_type = batch['attack_type'].to(self.device)
                target_model = batch['target_model'].to(self.device)
                labels = batch['labels'].to(self.device)
                
                # Forward pass
                outputs = self.model(input_ids, attack_type, target_model)
                
                # Calculate loss
                loss = criterion(outputs.view(-1, self.config.vocab_size), labels.view(-1))
                
                # Backward pass
                optimizer.zero_grad()
                loss.backward()
                optimizer.step()
                
                total_loss += loss.item()
            
            avg_loss = total_loss / len(train_loader)
            logger.info(f"Epoch {epoch+1}/{epochs}, Loss: {avg_loss:.4f}")
            
            # Validation
            if val_dataset:
                val_loss = self._validate(val_dataset, criterion)
                logger.info(f"Validation Loss: {val_loss:.4f}")
    
    def _prepare_dataset(self, data: List[Dict[str, Any]]) -> torch.utils.data.TensorDataset:
        """Prepare dataset for training"""
        all_input_ids = []
        all_labels = []
        all_attack_types = []
        all_target_models = []
        
        for item in data:
            # Encode payload
            tokens = self.tokenizer.encode(item['payload'], max_length=self.config.max_seq_length)
            
            # Create input and labels (shifted by 1 for next token prediction)
            input_ids = tokens[:-1]
            labels = tokens[1:]
            
            # Get attack type and target model
            attack_type = self.attack_types.get(item['attack_type'], 0)
            target_model = self.target_models.get(item['target_model'], 0)
            
            all_input_ids.append(input_ids)
            all_labels.append(labels)
            all_attack_types.append(attack_type)
            all_target_models.append(target_model)
        
        # Convert to tensors
        input_ids_tensor = torch.tensor(all_input_ids, dtype=torch.long)
        labels_tensor = torch.tensor(all_labels, dtype=torch.long)
        attack_types_tensor = torch.tensor(all_attack_types, dtype=torch.long)
        target_models_tensor = torch.tensor(all_target_models, dtype=torch.long)
        
        return torch.utils.data.TensorDataset(
            input_ids_tensor,
            labels_tensor,
            attack_types_tensor,
            target_models_tensor
        )
    
    def generate_attack(self,
                       attack_type: str,
                       target_model: str,
                       prompt: str = "",
                       max_length: int = 100,
                       temperature: float = 0.8,
                       top_k: int = 50,
                       top_p: float = 0.95,
                       num_samples: int = 1,
                       style_params: Optional[Dict[str, float]] = None) -> List[str]:
        """
        Generate attack payloads.
        
        Args:
            attack_type: Type of attack
            target_model: Target model name
            prompt: Starting prompt
            max_length: Maximum length
            temperature: Sampling temperature
            top_k: Top-k sampling
            top_p: Top-p sampling
            num_samples: Number of samples to generate
            style_params: Style control (aggression, formality, obfuscation, creativity)
            
        Returns:
            List of generated attack payloads
        """
        self.model.eval()
        
        # Prepare inputs
        if prompt:
            prompt_tokens = self.tokenizer.encode(prompt, max_length=50)
        else:
            prompt_tokens = [self.tokenizer.special_tokens['<START>']]
        
        # Create batch
        prompt_tensor = torch.tensor([prompt_tokens] * num_samples, dtype=torch.long).to(self.device)
        attack_type_id = self.attack_types.get(attack_type, 0)
        target_model_id = self.target_models.get(target_model, 0)
        
        attack_type_tensor = torch.tensor([attack_type_id] * num_samples, dtype=torch.long).to(self.device)
        target_model_tensor = torch.tensor([target_model_id] * num_samples, dtype=torch.long).to(self.device)
        
        # Style parameters
        if style_params:
            style_tensor = torch.tensor([
                [style_params.get('aggression', 0.5),
                 style_params.get('formality', 0.5),
                 style_params.get('obfuscation', 0.5),
                 style_params.get('creativity', 0.5)]
            ] * num_samples, dtype=torch.float).to(self.device)
        else:
            style_tensor = None
        
        # Generate
        generated = self.model.generate(
            prompt_tensor,
            attack_type_tensor,
            target_model_tensor,
            max_length=max_length,
            temperature=temperature,
            top_k=top_k,
            top_p=top_p,
            style_params=style_tensor
        )
        
        # Decode
        payloads = []
        for i in range(num_samples):
            tokens = generated[i].cpu().tolist()
            payload = self.tokenizer.decode(tokens)
            payloads.append(payload)
        
        return payloads
    
    def fine_tune_on_successful_attacks(self,
                                      successful_attacks: List[Dict[str, Any]],
                                      epochs: int = 5,
                                      learning_rate: float = 5e-5):
        """Fine-tune model on successful attacks"""
        logger.info(f"Fine-tuning on {len(successful_attacks)} successful attacks")
        
        # Filter only successful attacks
        successful_data = [
            attack for attack in successful_attacks
            if attack.get('success', False)
        ]
        
        if not successful_data:
            logger.warning("No successful attacks to fine-tune on")
            return
        
        # Train with lower learning rate
        self.train_on_dataset(
            successful_data,
            epochs=epochs,
            learning_rate=learning_rate
        )
    
    def _validate(self, val_dataset: torch.utils.data.TensorDataset, criterion: nn.Module) -> float:
        """Validate model"""
        self.model.eval()
        val_loader = torch.utils.data.DataLoader(val_dataset, batch_size=32)
        total_loss = 0
        
        with torch.no_grad():
            for batch in val_loader:
                input_ids = batch[0].to(self.device)
                labels = batch[1].to(self.device)
                attack_type = batch[2].to(self.device)
                target_model = batch[3].to(self.device)
                
                outputs = self.model(input_ids, attack_type, target_model)
                loss = criterion(outputs.view(-1, self.config.vocab_size), labels.view(-1))
                total_loss += loss.item()
        
        self.model.train()
        return total_loss / len(val_loader)
    
    def save_model(self, path: str):
        """Save model and tokenizer"""
        torch.save({
            'model_state_dict': self.model.state_dict(),
            'config': self.config,
            'tokenizer_vocab': self.tokenizer.word_to_idx,
            'attack_types': self.attack_types,
            'target_models': self.target_models
        }, path)
        
        logger.info(f"Model saved to {path}")
    
    def load_model(self, path: str):
        """Load model and tokenizer"""
        checkpoint = torch.load(path, map_location=self.device)
        
        self.config = checkpoint['config']
        self.model = AttackTransformer(self.config).to(self.device)
        self.model.load_state_dict(checkpoint['model_state_dict'])
        
        self.tokenizer.word_to_idx = checkpoint['tokenizer_vocab']
        self.tokenizer.idx_to_word = {v: k for k, v in self.tokenizer.word_to_idx.items()}
        
        self.attack_types = checkpoint['attack_types']
        self.target_models = checkpoint['target_models']
        
        logger.info(f"Model loaded from {path}")