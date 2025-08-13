"""
ML Model Storage Infrastructure for LLMrecon

This module provides a unified interface for storing, versioning, and
managing trained ML models across different storage backends (S3, local, distributed FS).
"""

import os
import json
import time
import hashlib
import pickle
import shutil
from abc import ABC, abstractmethod
from pathlib import Path
from typing import Dict, List, Optional, Any, Union, BinaryIO
from dataclasses import dataclass, asdict
from datetime import datetime
import threading
from concurrent.futures import ThreadPoolExecutor
import logging

# Optional imports for different backends
try:
    import boto3
    from botocore.exceptions import ClientError
    HAS_S3 = True
except ImportError:
    HAS_S3 = False

try:
    import torch
    HAS_TORCH = True
except ImportError:
    HAS_TORCH = False

try:
    import tensorflow as tf
    HAS_TF = True
except ImportError:
    HAS_TF = False

# Configure logging
logger = logging.getLogger(__name__)


@dataclass
class ModelMetadata:
    """Metadata for a stored model"""
    model_id: str
    name: str
    version: str
    type: str  # 'pytorch', 'tensorflow', 'sklearn', 'custom'
    algorithm: str  # 'dqn', 'genetic', 'transformer', etc.
    
    # Training info
    training_data_hash: Optional[str] = None
    training_params: Optional[Dict[str, Any]] = None
    performance_metrics: Optional[Dict[str, float]] = None
    
    # Storage info
    storage_path: Optional[str] = None
    file_size: Optional[int] = None
    checksum: Optional[str] = None
    
    # Timestamps
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None
    
    # Additional metadata
    tags: Optional[List[str]] = None
    description: Optional[str] = None
    author: Optional[str] = None
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        data = asdict(self)
        if self.created_at:
            data['created_at'] = self.created_at.isoformat()
        if self.updated_at:
            data['updated_at'] = self.updated_at.isoformat()
        return data
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'ModelMetadata':
        """Create from dictionary"""
        if 'created_at' in data and isinstance(data['created_at'], str):
            data['created_at'] = datetime.fromisoformat(data['created_at'])
        if 'updated_at' in data and isinstance(data['updated_at'], str):
            data['updated_at'] = datetime.fromisoformat(data['updated_at'])
        return cls(**data)


class StorageBackend(ABC):
    """Abstract base class for model storage backends"""
    
    @abstractmethod
    def save_model(self, model: Any, metadata: ModelMetadata) -> str:
        """Save a model and return storage path"""
        pass
    
    @abstractmethod
    def load_model(self, model_id: str, version: Optional[str] = None) -> Any:
        """Load a model by ID and optionally version"""
        pass
    
    @abstractmethod
    def list_models(self, **filters) -> List[ModelMetadata]:
        """List available models with optional filters"""
        pass
    
    @abstractmethod
    def delete_model(self, model_id: str, version: Optional[str] = None):
        """Delete a model"""
        pass
    
    @abstractmethod
    def get_metadata(self, model_id: str, version: Optional[str] = None) -> ModelMetadata:
        """Get model metadata"""
        pass


class LocalStorageBackend(StorageBackend):
    """Local filesystem storage backend"""
    
    def __init__(self, base_path: str = "models/storage"):
        self.base_path = Path(base_path)
        self.base_path.mkdir(parents=True, exist_ok=True)
        self.metadata_file = self.base_path / "metadata.json"
        self._load_metadata_index()
    
    def _load_metadata_index(self):
        """Load metadata index from disk"""
        if self.metadata_file.exists():
            with open(self.metadata_file, 'r') as f:
                data = json.load(f)
                self.metadata_index = {
                    k: ModelMetadata.from_dict(v) for k, v in data.items()
                }
        else:
            self.metadata_index = {}
    
    def _save_metadata_index(self):
        """Save metadata index to disk"""
        data = {k: v.to_dict() for k, v in self.metadata_index.items()}
        with open(self.metadata_file, 'w') as f:
            json.dump(data, f, indent=2)
    
    def _get_model_path(self, model_id: str, version: str) -> Path:
        """Get path for a specific model version"""
        return self.base_path / model_id / version
    
    def save_model(self, model: Any, metadata: ModelMetadata) -> str:
        """Save model to local filesystem"""
        # Create versioned directory
        model_path = self._get_model_path(metadata.model_id, metadata.version)
        model_path.mkdir(parents=True, exist_ok=True)
        
        # Determine save method based on model type
        model_file = model_path / "model.pkl"
        
        if metadata.type == "pytorch" and HAS_TORCH:
            model_file = model_path / "model.pt"
            torch.save(model, model_file)
        elif metadata.type == "tensorflow" and HAS_TF:
            model_file = model_path / "model"
            tf.saved_model.save(model, str(model_file))
        else:
            # Default to pickle for other types
            with open(model_file, 'wb') as f:
                pickle.dump(model, f)
        
        # Calculate checksum
        if model_file.is_file():
            with open(model_file, 'rb') as f:
                metadata.checksum = hashlib.sha256(f.read()).hexdigest()
            metadata.file_size = model_file.stat().st_size
        else:  # For TF saved models (directory)
            metadata.file_size = sum(
                f.stat().st_size for f in model_file.rglob('*') if f.is_file()
            )
        
        # Update metadata
        metadata.storage_path = str(model_file)
        metadata.created_at = datetime.now()
        metadata.updated_at = datetime.now()
        
        # Save metadata
        metadata_path = model_path / "metadata.json"
        with open(metadata_path, 'w') as f:
            json.dump(metadata.to_dict(), f, indent=2)
        
        # Update index
        key = f"{metadata.model_id}:{metadata.version}"
        self.metadata_index[key] = metadata
        self._save_metadata_index()
        
        logger.info(f"Saved model {metadata.model_id} v{metadata.version} to {model_path}")
        return str(model_file)
    
    def load_model(self, model_id: str, version: Optional[str] = None) -> Any:
        """Load model from local filesystem"""
        # Find latest version if not specified
        if version is None:
            versions = [
                k.split(':')[1] for k in self.metadata_index.keys()
                if k.startswith(f"{model_id}:")
            ]
            if not versions:
                raise ValueError(f"Model {model_id} not found")
            version = max(versions)  # Simple version comparison
        
        # Get metadata
        key = f"{model_id}:{version}"
        if key not in self.metadata_index:
            raise ValueError(f"Model {model_id} v{version} not found")
        
        metadata = self.metadata_index[key]
        model_path = Path(metadata.storage_path)
        
        # Load based on type
        if metadata.type == "pytorch" and HAS_TORCH:
            model = torch.load(model_path)
        elif metadata.type == "tensorflow" and HAS_TF:
            model = tf.saved_model.load(str(model_path))
        else:
            with open(model_path, 'rb') as f:
                model = pickle.load(f)
        
        logger.info(f"Loaded model {model_id} v{version}")
        return model
    
    def list_models(self, **filters) -> List[ModelMetadata]:
        """List models with optional filters"""
        models = list(self.metadata_index.values())
        
        # Apply filters
        if 'algorithm' in filters:
            models = [m for m in models if m.algorithm == filters['algorithm']]
        if 'type' in filters:
            models = [m for m in models if m.type == filters['type']]
        if 'tags' in filters:
            filter_tags = set(filters['tags'])
            models = [
                m for m in models 
                if m.tags and filter_tags.intersection(m.tags)
            ]
        
        return sorted(models, key=lambda m: (m.model_id, m.version))
    
    def delete_model(self, model_id: str, version: Optional[str] = None):
        """Delete model from local filesystem"""
        if version is None:
            # Delete all versions
            keys_to_delete = [
                k for k in self.metadata_index.keys()
                if k.startswith(f"{model_id}:")
            ]
        else:
            keys_to_delete = [f"{model_id}:{version}"]
        
        for key in keys_to_delete:
            if key in self.metadata_index:
                metadata = self.metadata_index[key]
                model_path = self._get_model_path(
                    metadata.model_id, metadata.version
                )
                
                # Remove directory
                if model_path.exists():
                    shutil.rmtree(model_path)
                
                # Remove from index
                del self.metadata_index[key]
                logger.info(f"Deleted model {key}")
        
        self._save_metadata_index()
    
    def get_metadata(self, model_id: str, version: Optional[str] = None) -> ModelMetadata:
        """Get model metadata"""
        if version is None:
            # Get latest version
            versions = [
                k.split(':')[1] for k in self.metadata_index.keys()
                if k.startswith(f"{model_id}:")
            ]
            if not versions:
                raise ValueError(f"Model {model_id} not found")
            version = max(versions)
        
        key = f"{model_id}:{version}"
        if key not in self.metadata_index:
            raise ValueError(f"Model {model_id} v{version} not found")
        
        return self.metadata_index[key]


class S3StorageBackend(StorageBackend):
    """AWS S3 storage backend"""
    
    def __init__(self, bucket_name: str, prefix: str = "models/"):
        if not HAS_S3:
            raise ImportError("boto3 is required for S3 storage backend")
        
        self.bucket_name = bucket_name
        self.prefix = prefix
        self.s3_client = boto3.client('s3')
        
        # Local cache for metadata
        self.cache_dir = Path(".cache/s3_models")
        self.cache_dir.mkdir(parents=True, exist_ok=True)
    
    def _get_s3_key(self, model_id: str, version: str, filename: str) -> str:
        """Get S3 key for a file"""
        return f"{self.prefix}{model_id}/{version}/{filename}"
    
    def save_model(self, model: Any, metadata: ModelMetadata) -> str:
        """Save model to S3"""
        # Save to temporary local file first
        temp_dir = self.cache_dir / f"temp_{metadata.model_id}_{metadata.version}"
        temp_dir.mkdir(parents=True, exist_ok=True)
        
        try:
            # Save model locally
            local_backend = LocalStorageBackend(str(temp_dir))
            local_path = local_backend.save_model(model, metadata)
            
            # Upload to S3
            model_key = self._get_s3_key(
                metadata.model_id, 
                metadata.version, 
                "model.pkl"
            )
            
            # Determine actual file to upload
            local_file = Path(local_path)
            if local_file.is_dir():  # TensorFlow saved model
                # Create tar archive
                import tarfile
                tar_path = temp_dir / "model.tar.gz"
                with tarfile.open(tar_path, "w:gz") as tar:
                    tar.add(local_file, arcname="model")
                local_file = tar_path
                model_key = self._get_s3_key(
                    metadata.model_id,
                    metadata.version,
                    "model.tar.gz"
                )
            
            # Upload model
            self.s3_client.upload_file(
                str(local_file),
                self.bucket_name,
                model_key
            )
            
            # Upload metadata
            metadata_key = self._get_s3_key(
                metadata.model_id,
                metadata.version,
                "metadata.json"
            )
            self.s3_client.put_object(
                Bucket=self.bucket_name,
                Key=metadata_key,
                Body=json.dumps(metadata.to_dict()).encode('utf-8')
            )
            
            # Update storage path
            metadata.storage_path = f"s3://{self.bucket_name}/{model_key}"
            
            logger.info(f"Saved model {metadata.model_id} v{metadata.version} to S3")
            return metadata.storage_path
            
        finally:
            # Cleanup temp directory
            if temp_dir.exists():
                shutil.rmtree(temp_dir)
    
    def load_model(self, model_id: str, version: Optional[str] = None) -> Any:
        """Load model from S3"""
        # Get metadata first to determine version
        metadata = self.get_metadata(model_id, version)
        version = metadata.version
        
        # Check cache first
        cache_path = self.cache_dir / model_id / version
        if cache_path.exists():
            local_backend = LocalStorageBackend(str(self.cache_dir))
            return local_backend.load_model(model_id, version)
        
        # Download from S3
        cache_path.mkdir(parents=True, exist_ok=True)
        
        # Download metadata
        metadata_key = self._get_s3_key(model_id, version, "metadata.json")
        metadata_path = cache_path / "metadata.json"
        self.s3_client.download_file(
            self.bucket_name,
            metadata_key,
            str(metadata_path)
        )
        
        # Determine model file name
        model_file = "model.pkl"
        if metadata.type == "pytorch":
            model_file = "model.pt"
        elif metadata.type == "tensorflow":
            model_file = "model.tar.gz"
        
        # Download model
        model_key = self._get_s3_key(model_id, version, model_file)
        local_model_path = cache_path / model_file
        self.s3_client.download_file(
            self.bucket_name,
            model_key,
            str(local_model_path)
        )
        
        # Extract if needed (TensorFlow)
        if model_file == "model.tar.gz":
            import tarfile
            with tarfile.open(local_model_path, "r:gz") as tar:
                tar.extractall(cache_path)
            local_model_path = cache_path / "model"
        
        # Update metadata with local path
        metadata.storage_path = str(local_model_path)
        with open(metadata_path, 'w') as f:
            json.dump(metadata.to_dict(), f)
        
        # Load using local backend
        local_backend = LocalStorageBackend(str(self.cache_dir))
        return local_backend.load_model(model_id, version)
    
    def list_models(self, **filters) -> List[ModelMetadata]:
        """List models in S3"""
        models = []
        
        # List all metadata files
        paginator = self.s3_client.get_paginator('list_objects_v2')
        pages = paginator.paginate(
            Bucket=self.bucket_name,
            Prefix=self.prefix,
            Delimiter='/'
        )
        
        for page in pages:
            if 'Contents' in page:
                for obj in page['Contents']:
                    if obj['Key'].endswith('/metadata.json'):
                        # Download and parse metadata
                        try:
                            response = self.s3_client.get_object(
                                Bucket=self.bucket_name,
                                Key=obj['Key']
                            )
                            metadata_dict = json.loads(
                                response['Body'].read().decode('utf-8')
                            )
                            metadata = ModelMetadata.from_dict(metadata_dict)
                            
                            # Apply filters
                            if self._matches_filters(metadata, filters):
                                models.append(metadata)
                                
                        except Exception as e:
                            logger.error(f"Error loading metadata from {obj['Key']}: {e}")
        
        return sorted(models, key=lambda m: (m.model_id, m.version))
    
    def _matches_filters(self, metadata: ModelMetadata, filters: Dict[str, Any]) -> bool:
        """Check if metadata matches filters"""
        if 'algorithm' in filters and metadata.algorithm != filters['algorithm']:
            return False
        if 'type' in filters and metadata.type != filters['type']:
            return False
        if 'tags' in filters:
            filter_tags = set(filters['tags'])
            if not metadata.tags or not filter_tags.intersection(metadata.tags):
                return False
        return True
    
    def delete_model(self, model_id: str, version: Optional[str] = None):
        """Delete model from S3"""
        if version is None:
            # List all versions
            prefix = f"{self.prefix}{model_id}/"
            paginator = self.s3_client.get_paginator('list_objects_v2')
            pages = paginator.paginate(Bucket=self.bucket_name, Prefix=prefix)
            
            objects_to_delete = []
            for page in pages:
                if 'Contents' in page:
                    objects_to_delete.extend([
                        {'Key': obj['Key']} for obj in page['Contents']
                    ])
        else:
            # Delete specific version
            prefix = f"{self.prefix}{model_id}/{version}/"
            response = self.s3_client.list_objects_v2(
                Bucket=self.bucket_name,
                Prefix=prefix
            )
            
            if 'Contents' in response:
                objects_to_delete = [
                    {'Key': obj['Key']} for obj in response['Contents']
                ]
            else:
                objects_to_delete = []
        
        # Delete objects
        if objects_to_delete:
            self.s3_client.delete_objects(
                Bucket=self.bucket_name,
                Delete={'Objects': objects_to_delete}
            )
            logger.info(f"Deleted {len(objects_to_delete)} objects for {model_id}")
        
        # Clear cache
        cache_path = self.cache_dir / model_id
        if version:
            cache_path = cache_path / version
        if cache_path.exists():
            shutil.rmtree(cache_path)
    
    def get_metadata(self, model_id: str, version: Optional[str] = None) -> ModelMetadata:
        """Get model metadata from S3"""
        if version is None:
            # Find latest version
            prefix = f"{self.prefix}{model_id}/"
            response = self.s3_client.list_objects_v2(
                Bucket=self.bucket_name,
                Prefix=prefix,
                Delimiter='/'
            )
            
            if 'CommonPrefixes' not in response:
                raise ValueError(f"Model {model_id} not found")
            
            versions = [
                p['Prefix'].rstrip('/').split('/')[-1]
                for p in response['CommonPrefixes']
            ]
            version = max(versions)
        
        # Download metadata
        metadata_key = self._get_s3_key(model_id, version, "metadata.json")
        
        try:
            response = self.s3_client.get_object(
                Bucket=self.bucket_name,
                Key=metadata_key
            )
            metadata_dict = json.loads(response['Body'].read().decode('utf-8'))
            return ModelMetadata.from_dict(metadata_dict)
            
        except ClientError as e:
            if e.response['Error']['Code'] == 'NoSuchKey':
                raise ValueError(f"Model {model_id} v{version} not found")
            raise


class ModelRegistry:
    """
    Unified model registry that manages multiple storage backends
    and provides versioning, tagging, and model lifecycle management.
    """
    
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.backends = {}
        
        # Initialize backends based on config
        if config.get('local', {}).get('enabled', True):
            self.backends['local'] = LocalStorageBackend(
                config['local'].get('path', 'models/storage')
            )
        
        if config.get('s3', {}).get('enabled', False):
            self.backends['s3'] = S3StorageBackend(
                config['s3']['bucket'],
                config['s3'].get('prefix', 'models/')
            )
        
        # Set primary backend
        self.primary_backend = config.get('primary', 'local')
        if self.primary_backend not in self.backends:
            raise ValueError(f"Primary backend {self.primary_backend} not configured")
        
        # Model version tracking
        self.version_lock = threading.Lock()
        
    def save_model(self,
                   model: Any,
                   name: str,
                   algorithm: str,
                   model_type: str = 'custom',
                   version: Optional[str] = None,
                   tags: Optional[List[str]] = None,
                   description: Optional[str] = None,
                   training_params: Optional[Dict[str, Any]] = None,
                   performance_metrics: Optional[Dict[str, float]] = None,
                   author: Optional[str] = None) -> ModelMetadata:
        """
        Save a model to the registry.
        
        Args:
            model: The model object to save
            name: Human-readable name for the model
            algorithm: Algorithm type (dqn, genetic, transformer, etc.)
            model_type: Framework type (pytorch, tensorflow, sklearn, custom)
            version: Version string (auto-generated if not provided)
            tags: List of tags for categorization
            description: Model description
            training_params: Parameters used for training
            performance_metrics: Performance metrics from evaluation
            author: Model author
            
        Returns:
            ModelMetadata object with storage details
        """
        # Generate model ID
        model_id = f"{algorithm}_{name}".lower().replace(' ', '_')
        
        # Auto-generate version if needed
        if version is None:
            with self.version_lock:
                existing_versions = self._get_versions(model_id)
                if existing_versions:
                    # Increment patch version
                    latest = max(existing_versions)
                    parts = latest.split('.')
                    if len(parts) == 3 and parts[2].isdigit():
                        parts[2] = str(int(parts[2]) + 1)
                        version = '.'.join(parts)
                    else:
                        version = f"{latest}.1"
                else:
                    version = "1.0.0"
        
        # Create metadata
        metadata = ModelMetadata(
            model_id=model_id,
            name=name,
            version=version,
            type=model_type,
            algorithm=algorithm,
            tags=tags or [],
            description=description,
            training_params=training_params,
            performance_metrics=performance_metrics,
            author=author or "unknown"
        )
        
        # Calculate training data hash if provided
        if training_params and 'training_data_path' in training_params:
            metadata.training_data_hash = self._hash_file(
                training_params['training_data_path']
            )
        
        # Save to primary backend
        primary = self.backends[self.primary_backend]
        storage_path = primary.save_model(model, metadata)
        
        # Replicate to other backends asynchronously
        if len(self.backends) > 1:
            self._replicate_async(model, metadata)
        
        logger.info(f"Saved model {model_id} v{version} to registry")
        return metadata
    
    def load_model(self,
                   model_id: str,
                   version: Optional[str] = None,
                   backend: Optional[str] = None) -> Any:
        """
        Load a model from the registry.
        
        Args:
            model_id: Model identifier
            version: Specific version (latest if not provided)
            backend: Specific backend to use (primary if not provided)
            
        Returns:
            The loaded model object
        """
        backend_name = backend or self.primary_backend
        if backend_name not in self.backends:
            raise ValueError(f"Backend {backend_name} not available")
        
        return self.backends[backend_name].load_model(model_id, version)
    
    def list_models(self,
                    algorithm: Optional[str] = None,
                    tags: Optional[List[str]] = None,
                    model_type: Optional[str] = None) -> List[ModelMetadata]:
        """List available models with optional filters"""
        filters = {}
        if algorithm:
            filters['algorithm'] = algorithm
        if tags:
            filters['tags'] = tags
        if model_type:
            filters['type'] = model_type
        
        # Get from primary backend
        return self.backends[self.primary_backend].list_models(**filters)
    
    def get_metadata(self,
                     model_id: str,
                     version: Optional[str] = None) -> ModelMetadata:
        """Get model metadata"""
        return self.backends[self.primary_backend].get_metadata(model_id, version)
    
    def delete_model(self,
                     model_id: str,
                     version: Optional[str] = None,
                     all_backends: bool = False):
        """
        Delete a model from the registry.
        
        Args:
            model_id: Model identifier
            version: Specific version (all versions if not provided)
            all_backends: Delete from all backends (default: primary only)
        """
        if all_backends:
            for backend in self.backends.values():
                backend.delete_model(model_id, version)
        else:
            self.backends[self.primary_backend].delete_model(model_id, version)
    
    def compare_models(self,
                       model_ids: List[str],
                       metric: str = 'success_rate') -> Dict[str, Any]:
        """Compare performance metrics across models"""
        comparison = {}
        
        for model_id in model_ids:
            try:
                metadata = self.get_metadata(model_id)
                if metadata.performance_metrics and metric in metadata.performance_metrics:
                    comparison[model_id] = {
                        'version': metadata.version,
                        'value': metadata.performance_metrics[metric],
                        'algorithm': metadata.algorithm,
                        'created_at': metadata.created_at
                    }
            except ValueError:
                logger.warning(f"Model {model_id} not found")
        
        return comparison
    
    def _get_versions(self, model_id: str) -> List[str]:
        """Get all versions of a model"""
        models = self.list_models()
        return [m.version for m in models if m.model_id == model_id]
    
    def _hash_file(self, file_path: str) -> str:
        """Calculate SHA256 hash of a file"""
        sha256_hash = hashlib.sha256()
        with open(file_path, "rb") as f:
            for byte_block in iter(lambda: f.read(4096), b""):
                sha256_hash.update(byte_block)
        return sha256_hash.hexdigest()
    
    def _replicate_async(self, model: Any, metadata: ModelMetadata):
        """Asynchronously replicate model to other backends"""
        def replicate():
            for name, backend in self.backends.items():
                if name != self.primary_backend:
                    try:
                        backend.save_model(model, metadata)
                        logger.info(f"Replicated {metadata.model_id} to {name}")
                    except Exception as e:
                        logger.error(f"Failed to replicate to {name}: {e}")
        
        thread = threading.Thread(target=replicate)
        thread.daemon = True
        thread.start()


# Example usage
def create_model_registry(config: Optional[Dict[str, Any]] = None) -> ModelRegistry:
    """Create and configure model registry"""
    default_config = {
        'primary': 'local',
        'local': {
            'enabled': True,
            'path': 'ml/models'
        },
        's3': {
            'enabled': False,
            'bucket': 'llmrecon-models',
            'prefix': 'models/'
        }
    }
    
    if config:
        default_config.update(config)
    
    return ModelRegistry(default_config)