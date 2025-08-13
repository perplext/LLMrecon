"""
Model Versioning and Lifecycle Management for LLMrecon

This module provides advanced versioning capabilities including semantic versioning,
model lineage tracking, and automated lifecycle management.
"""

import re
import json
from typing import Dict, List, Optional, Tuple, Any
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum
import logging

from .model_storage import ModelMetadata, ModelRegistry

logger = logging.getLogger(__name__)


class ModelStage(Enum):
    """Model lifecycle stages"""
    DEVELOPMENT = "development"
    STAGING = "staging"
    PRODUCTION = "production"
    ARCHIVED = "archived"


@dataclass
class ModelVersion:
    """Semantic version representation"""
    major: int
    minor: int
    patch: int
    prerelease: Optional[str] = None
    build: Optional[str] = None
    
    def __str__(self) -> str:
        version = f"{self.major}.{self.minor}.{self.patch}"
        if self.prerelease:
            version += f"-{self.prerelease}"
        if self.build:
            version += f"+{self.build}"
        return version
    
    @classmethod
    def parse(cls, version_str: str) -> 'ModelVersion':
        """Parse semantic version string"""
        pattern = r'^(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z0-9.-]+))?(?:\+([a-zA-Z0-9.-]+))?$'
        match = re.match(pattern, version_str)
        
        if not match:
            raise ValueError(f"Invalid version format: {version_str}")
        
        return cls(
            major=int(match.group(1)),
            minor=int(match.group(2)),
            patch=int(match.group(3)),
            prerelease=match.group(4),
            build=match.group(5)
        )
    
    def bump(self, component: str = 'patch') -> 'ModelVersion':
        """Bump version component"""
        if component == 'major':
            return ModelVersion(self.major + 1, 0, 0)
        elif component == 'minor':
            return ModelVersion(self.major, self.minor + 1, 0)
        elif component == 'patch':
            return ModelVersion(self.major, self.minor, self.patch + 1)
        else:
            raise ValueError(f"Unknown component: {component}")
    
    def __lt__(self, other: 'ModelVersion') -> bool:
        """Compare versions"""
        # Compare main version numbers
        if (self.major, self.minor, self.patch) != (other.major, other.minor, other.patch):
            return (self.major, self.minor, self.patch) < (other.major, other.minor, other.patch)
        
        # Handle prerelease versions
        if self.prerelease and not other.prerelease:
            return True  # Prerelease < release
        elif not self.prerelease and other.prerelease:
            return False  # Release > prerelease
        elif self.prerelease and other.prerelease:
            return self.prerelease < other.prerelease
        
        return False


@dataclass
class ModelLineage:
    """Track model lineage and relationships"""
    model_id: str
    version: str
    parent_model_id: Optional[str] = None
    parent_version: Optional[str] = None
    children: List[Tuple[str, str]] = field(default_factory=list)
    experiment_id: Optional[str] = None
    created_from: Optional[str] = None  # 'training', 'fine-tuning', 'transfer-learning'
    
    def add_child(self, child_model_id: str, child_version: str):
        """Add a child model"""
        self.children.append((child_model_id, child_version))


class ModelVersionManager:
    """
    Manages model versions, transitions, and lifecycle.
    
    Features:
    - Semantic versioning with automatic bumping
    - Stage transitions (dev -> staging -> prod)
    - Model lineage tracking
    - Automated cleanup policies
    - Version comparison and rollback
    """
    
    def __init__(self, registry: ModelRegistry):
        self.registry = registry
        self.stages_file = "ml/models/stages.json"
        self.lineage_file = "ml/models/lineage.json"
        self._load_state()
    
    def _load_state(self):
        """Load stages and lineage from disk"""
        # Load stages
        try:
            with open(self.stages_file, 'r') as f:
                data = json.load(f)
                self.stages = {
                    k: ModelStage(v) for k, v in data.items()
                }
        except FileNotFoundError:
            self.stages = {}
        
        # Load lineage
        try:
            with open(self.lineage_file, 'r') as f:
                data = json.load(f)
                self.lineage = {
                    k: ModelLineage(**v) for k, v in data.items()
                }
        except FileNotFoundError:
            self.lineage = {}
    
    def _save_state(self):
        """Save stages and lineage to disk"""
        # Save stages
        with open(self.stages_file, 'w') as f:
            data = {k: v.value for k, v in self.stages.items()}
            json.dump(data, f, indent=2)
        
        # Save lineage
        with open(self.lineage_file, 'w') as f:
            data = {k: v.__dict__ for k, v in self.lineage.items()}
            json.dump(data, f, indent=2)
    
    def register_model(self,
                      model: Any,
                      name: str,
                      algorithm: str,
                      parent_model: Optional[Tuple[str, str]] = None,
                      bump_type: str = 'patch',
                      stage: ModelStage = ModelStage.DEVELOPMENT,
                      **kwargs) -> ModelMetadata:
        """
        Register a new model version with automatic versioning.
        
        Args:
            model: Model object
            name: Model name
            algorithm: Algorithm type
            parent_model: Parent (model_id, version) if derived
            bump_type: Version bump type (major, minor, patch)
            stage: Initial stage
            **kwargs: Additional metadata
            
        Returns:
            ModelMetadata with registration details
        """
        model_id = f"{algorithm}_{name}".lower().replace(' ', '_')
        
        # Determine version
        if parent_model:
            parent_id, parent_version = parent_model
            parent_ver = ModelVersion.parse(parent_version)
            new_version = parent_ver.bump(bump_type)
        else:
            # Check for existing versions
            existing_versions = self._get_model_versions(model_id)
            if existing_versions:
                latest = max(existing_versions, key=lambda v: ModelVersion.parse(v))
                latest_ver = ModelVersion.parse(latest)
                new_version = latest_ver.bump(bump_type)
            else:
                new_version = ModelVersion(1, 0, 0)
        
        # Save model
        metadata = self.registry.save_model(
            model=model,
            name=name,
            algorithm=algorithm,
            version=str(new_version),
            **kwargs
        )
        
        # Update stage
        version_key = f"{model_id}:{new_version}"
        self.stages[version_key] = stage
        
        # Update lineage
        lineage = ModelLineage(
            model_id=model_id,
            version=str(new_version),
            parent_model_id=parent_model[0] if parent_model else None,
            parent_version=parent_model[1] if parent_model else None,
            experiment_id=kwargs.get('experiment_id'),
            created_from=kwargs.get('created_from', 'training')
        )
        
        # Update parent's children
        if parent_model:
            parent_key = f"{parent_model[0]}:{parent_model[1]}"
            if parent_key in self.lineage:
                self.lineage[parent_key].add_child(model_id, str(new_version))
        
        self.lineage[version_key] = lineage
        self._save_state()
        
        logger.info(f"Registered model {model_id} v{new_version} in {stage.value}")
        return metadata
    
    def transition_stage(self,
                        model_id: str,
                        version: str,
                        new_stage: ModelStage,
                        force: bool = False) -> bool:
        """
        Transition model to a new stage.
        
        Args:
            model_id: Model identifier
            version: Model version
            new_stage: Target stage
            force: Force transition (skip validation)
            
        Returns:
            Success status
        """
        version_key = f"{model_id}:{version}"
        current_stage = self.stages.get(version_key, ModelStage.DEVELOPMENT)
        
        # Validate transition
        if not force and not self._validate_transition(current_stage, new_stage):
            logger.error(f"Invalid transition from {current_stage} to {new_stage}")
            return False
        
        # Special handling for production
        if new_stage == ModelStage.PRODUCTION:
            # Archive previous production model
            self._archive_production_models(model_id)
        
        # Update stage
        self.stages[version_key] = new_stage
        self._save_state()
        
        logger.info(f"Transitioned {model_id} v{version} to {new_stage.value}")
        return True
    
    def _validate_transition(self,
                           current: ModelStage,
                           target: ModelStage) -> bool:
        """Validate stage transition"""
        valid_transitions = {
            ModelStage.DEVELOPMENT: [ModelStage.STAGING, ModelStage.ARCHIVED],
            ModelStage.STAGING: [ModelStage.PRODUCTION, ModelStage.DEVELOPMENT, ModelStage.ARCHIVED],
            ModelStage.PRODUCTION: [ModelStage.ARCHIVED],
            ModelStage.ARCHIVED: [ModelStage.DEVELOPMENT]  # Allow revival
        }
        
        return target in valid_transitions.get(current, [])
    
    def _archive_production_models(self, model_id: str):
        """Archive existing production models"""
        for version_key, stage in self.stages.items():
            if version_key.startswith(f"{model_id}:") and stage == ModelStage.PRODUCTION:
                self.stages[version_key] = ModelStage.ARCHIVED
                logger.info(f"Archived {version_key}")
    
    def get_model_by_stage(self,
                          model_id: str,
                          stage: ModelStage) -> Optional[ModelMetadata]:
        """Get model in specific stage"""
        for version_key, model_stage in self.stages.items():
            if version_key.startswith(f"{model_id}:") and model_stage == stage:
                _, version = version_key.split(':')
                return self.registry.get_metadata(model_id, version)
        return None
    
    def get_production_model(self, model_id: str) -> Optional[Any]:
        """Get current production model"""
        metadata = self.get_model_by_stage(model_id, ModelStage.PRODUCTION)
        if metadata:
            return self.registry.load_model(model_id, metadata.version)
        return None
    
    def get_lineage(self, model_id: str, version: str) -> Dict[str, Any]:
        """Get complete lineage tree for a model"""
        version_key = f"{model_id}:{version}"
        if version_key not in self.lineage:
            return {}
        
        lineage = self.lineage[version_key]
        
        result = {
            'model_id': model_id,
            'version': version,
            'stage': self.stages.get(version_key, ModelStage.DEVELOPMENT).value,
            'parent': None,
            'children': []
        }
        
        # Add parent info
        if lineage.parent_model_id:
            parent_key = f"{lineage.parent_model_id}:{lineage.parent_version}"
            parent_stage = self.stages.get(parent_key, ModelStage.ARCHIVED)
            result['parent'] = {
                'model_id': lineage.parent_model_id,
                'version': lineage.parent_version,
                'stage': parent_stage.value
            }
        
        # Add children info
        for child_id, child_version in lineage.children:
            child_key = f"{child_id}:{child_version}"
            child_stage = self.stages.get(child_key, ModelStage.DEVELOPMENT)
            result['children'].append({
                'model_id': child_id,
                'version': child_version,
                'stage': child_stage.value
            })
        
        return result
    
    def compare_versions(self,
                        model_id: str,
                        version1: str,
                        version2: str,
                        metric: str = 'success_rate') -> Dict[str, Any]:
        """Compare two model versions"""
        metadata1 = self.registry.get_metadata(model_id, version1)
        metadata2 = self.registry.get_metadata(model_id, version2)
        
        comparison = {
            'version1': {
                'version': version1,
                'stage': self.stages.get(f"{model_id}:{version1}", ModelStage.DEVELOPMENT).value,
                'created_at': metadata1.created_at,
                'metrics': metadata1.performance_metrics or {}
            },
            'version2': {
                'version': version2,
                'stage': self.stages.get(f"{model_id}:{version2}", ModelStage.DEVELOPMENT).value,
                'created_at': metadata2.created_at,
                'metrics': metadata2.performance_metrics or {}
            }
        }
        
        # Calculate improvement
        if metric in metadata1.performance_metrics and metric in metadata2.performance_metrics:
            val1 = metadata1.performance_metrics[metric]
            val2 = metadata2.performance_metrics[metric]
            improvement = ((val2 - val1) / val1) * 100 if val1 > 0 else 0
            comparison['improvement'] = {
                'metric': metric,
                'change': improvement,
                'direction': 'better' if improvement > 0 else 'worse'
            }
        
        return comparison
    
    def rollback_production(self, model_id: str) -> bool:
        """Rollback to previous production version"""
        # Find current production
        current_prod = None
        for version_key, stage in self.stages.items():
            if version_key.startswith(f"{model_id}:") and stage == ModelStage.PRODUCTION:
                current_prod = version_key
                break
        
        if not current_prod:
            logger.error(f"No production model found for {model_id}")
            return False
        
        # Find most recent archived version
        archived_versions = []
        for version_key, stage in self.stages.items():
            if version_key.startswith(f"{model_id}:") and stage == ModelStage.ARCHIVED:
                _, version = version_key.split(':')
                archived_versions.append((version_key, ModelVersion.parse(version)))
        
        if not archived_versions:
            logger.error(f"No archived versions found for {model_id}")
            return False
        
        # Sort by version and get latest
        archived_versions.sort(key=lambda x: x[1], reverse=True)
        rollback_key, _ = archived_versions[0]
        
        # Perform rollback
        self.stages[current_prod] = ModelStage.ARCHIVED
        self.stages[rollback_key] = ModelStage.PRODUCTION
        self._save_state()
        
        logger.info(f"Rolled back {model_id} from {current_prod} to {rollback_key}")
        return True
    
    def cleanup_old_versions(self,
                           model_id: str,
                           keep_last_n: int = 5,
                           keep_production: bool = True,
                           keep_staging: bool = True) -> List[str]:
        """
        Clean up old model versions.
        
        Args:
            model_id: Model identifier
            keep_last_n: Number of recent versions to keep
            keep_production: Keep production versions
            keep_staging: Keep staging versions
            
        Returns:
            List of deleted versions
        """
        # Get all versions
        all_versions = []
        for version_key in self.stages.keys():
            if version_key.startswith(f"{model_id}:"):
                _, version = version_key.split(':')
                stage = self.stages[version_key]
                all_versions.append((version, stage, ModelVersion.parse(version)))
        
        # Sort by version
        all_versions.sort(key=lambda x: x[2], reverse=True)
        
        # Determine which to keep
        keep_versions = set()
        deleted_versions = []
        
        # Keep recent versions
        for i, (version, stage, _) in enumerate(all_versions):
            if i < keep_last_n:
                keep_versions.add(version)
            elif keep_production and stage == ModelStage.PRODUCTION:
                keep_versions.add(version)
            elif keep_staging and stage == ModelStage.STAGING:
                keep_versions.add(version)
            else:
                deleted_versions.append(version)
        
        # Delete old versions
        for version in deleted_versions:
            self.registry.delete_model(model_id, version)
            version_key = f"{model_id}:{version}"
            if version_key in self.stages:
                del self.stages[version_key]
            if version_key in self.lineage:
                del self.lineage[version_key]
        
        self._save_state()
        logger.info(f"Cleaned up {len(deleted_versions)} versions of {model_id}")
        
        return deleted_versions
    
    def _get_model_versions(self, model_id: str) -> List[str]:
        """Get all versions of a model"""
        models = self.registry.list_models()
        return [m.version for m in models if m.model_id == model_id]
    
    def export_model_history(self, model_id: str) -> Dict[str, Any]:
        """Export complete model history"""
        versions = self._get_model_versions(model_id)
        
        history = {
            'model_id': model_id,
            'total_versions': len(versions),
            'versions': []
        }
        
        for version in sorted(versions, key=lambda v: ModelVersion.parse(v), reverse=True):
            version_key = f"{model_id}:{version}"
            metadata = self.registry.get_metadata(model_id, version)
            
            version_info = {
                'version': version,
                'stage': self.stages.get(version_key, ModelStage.DEVELOPMENT).value,
                'created_at': metadata.created_at.isoformat() if metadata.created_at else None,
                'algorithm': metadata.algorithm,
                'performance_metrics': metadata.performance_metrics,
                'lineage': self.lineage.get(version_key, {}).__dict__ if version_key in self.lineage else {}
            }
            
            history['versions'].append(version_info)
        
        return history