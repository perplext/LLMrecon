"""
Attack Data Pipeline for LLMrecon

This module handles collection, preprocessing, and storage of attack
outcomes for ML training. It integrates with the main application to
capture real attack data.
"""

import json
import time
import hashlib
import sqlite3
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Any, Optional, Tuple
from dataclasses import dataclass, asdict, field
from enum import Enum
import numpy as np
import pandas as pd
from collections import defaultdict
import threading
from queue import Queue
import logging


# Configure logging
logger = logging.getLogger(__name__)


class AttackStatus(Enum):
    """Status of an attack attempt"""
    SUCCESS = "success"
    FAILED = "failed"
    DETECTED = "detected"
    TIMEOUT = "timeout"
    ERROR = "error"


@dataclass
class AttackData:
    """Structured data for a single attack attempt"""
    # Identifiers
    attack_id: str
    timestamp: datetime
    
    # Attack parameters
    attack_type: str
    target_model: str
    provider: str
    
    # Attack details
    payload: str
    technique_params: Dict[str, Any]
    obfuscation_level: float
    
    # Outcomes
    status: AttackStatus
    response: str
    response_time: float
    tokens_used: int
    
    # Analysis
    success_indicators: List[str] = field(default_factory=list)
    detection_score: float = 0.0
    semantic_similarity: float = 0.0
    
    # Metadata
    session_id: Optional[str] = None
    user_id: Optional[str] = None
    campaign_id: Optional[str] = None
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for storage"""
        data = asdict(self)
        data['timestamp'] = self.timestamp.isoformat()
        data['status'] = self.status.value
        return data
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'AttackData':
        """Create from dictionary"""
        data['timestamp'] = datetime.fromisoformat(data['timestamp'])
        data['status'] = AttackStatus(data['status'])
        return cls(**data)


class AttackDataPipeline:
    """
    Pipeline for collecting and processing attack data.
    
    Features:
    - Real-time data collection
    - Preprocessing and feature extraction
    - Data validation and cleaning
    - Storage in multiple formats
    - Integration with ML training
    """
    
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.storage_path = Path(config.get('storage_path', 'data/attacks'))
        self.storage_path.mkdir(parents=True, exist_ok=True)
        
        # Database setup
        self.db_path = self.storage_path / 'attacks.db'
        self._init_database()
        
        # Data queues
        self.data_queue = Queue(maxsize=10000)
        self.processed_queue = Queue(maxsize=5000)
        
        # Statistics
        self.stats = defaultdict(int)
        self.start_time = time.time()
        
        # Processing thread
        self.processing_thread = threading.Thread(target=self._process_data_loop)
        self.processing_thread.daemon = True
        self.processing_thread.start()
        
        # Feature extractors
        self.feature_extractors = self._init_feature_extractors()
        
    def _init_database(self):
        """Initialize SQLite database for attack data"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS attacks (
                attack_id TEXT PRIMARY KEY,
                timestamp TEXT NOT NULL,
                attack_type TEXT NOT NULL,
                target_model TEXT NOT NULL,
                provider TEXT NOT NULL,
                payload TEXT NOT NULL,
                technique_params TEXT,
                obfuscation_level REAL,
                status TEXT NOT NULL,
                response TEXT,
                response_time REAL,
                tokens_used INTEGER,
                success_indicators TEXT,
                detection_score REAL,
                semantic_similarity REAL,
                session_id TEXT,
                user_id TEXT,
                campaign_id TEXT,
                features TEXT,
                created_at TEXT DEFAULT CURRENT_TIMESTAMP
            )
        ''')
        
        # Create indices for common queries
        cursor.execute('CREATE INDEX IF NOT EXISTS idx_timestamp ON attacks(timestamp)')
        cursor.execute('CREATE INDEX IF NOT EXISTS idx_attack_type ON attacks(attack_type)')
        cursor.execute('CREATE INDEX IF NOT EXISTS idx_status ON attacks(status)')
        cursor.execute('CREATE INDEX IF NOT EXISTS idx_target_model ON attacks(target_model)')
        
        conn.commit()
        conn.close()
        
    def _init_feature_extractors(self) -> Dict[str, Any]:
        """Initialize feature extraction functions"""
        return {
            'payload_features': self._extract_payload_features,
            'response_features': self._extract_response_features,
            'temporal_features': self._extract_temporal_features,
            'statistical_features': self._extract_statistical_features
        }
    
    def collect(self, attack_data: AttackData):
        """
        Collect new attack data.
        
        This is the main entry point for the data pipeline.
        """
        try:
            # Validate data
            if not self._validate_attack_data(attack_data):
                logger.warning(f"Invalid attack data: {attack_data.attack_id}")
                return
            
            # Add to queue for processing
            self.data_queue.put(attack_data)
            self.stats['collected'] += 1
            
        except Exception as e:
            logger.error(f"Error collecting attack data: {e}")
            self.stats['collection_errors'] += 1
    
    def _validate_attack_data(self, attack_data: AttackData) -> bool:
        """Validate attack data before processing"""
        # Check required fields
        if not attack_data.attack_id or not attack_data.payload:
            return False
            
        # Check payload length
        if len(attack_data.payload) > 10000:
            logger.warning(f"Payload too long: {len(attack_data.payload)}")
            return False
            
        # Check response time
        if attack_data.response_time < 0 or attack_data.response_time > 300:
            logger.warning(f"Invalid response time: {attack_data.response_time}")
            return False
            
        return True
    
    def _process_data_loop(self):
        """Background thread for processing attack data"""
        while True:
            try:
                # Get data from queue
                attack_data = self.data_queue.get(timeout=1.0)
                
                # Process and extract features
                processed_data = self._process_attack_data(attack_data)
                
                # Store in database
                self._store_attack_data(processed_data)
                
                # Add to processed queue for ML training
                self.processed_queue.put(processed_data)
                
                self.stats['processed'] += 1
                
            except Exception as e:
                if str(e) != "":  # Ignore empty timeout exceptions
                    logger.error(f"Error processing data: {e}")
                    self.stats['processing_errors'] += 1
    
    def _process_attack_data(self, attack_data: AttackData) -> Dict[str, Any]:
        """Process attack data and extract features"""
        # Convert to dict
        data_dict = attack_data.to_dict()
        
        # Extract features
        features = {}
        for name, extractor in self.feature_extractors.items():
            try:
                features[name] = extractor(attack_data)
            except Exception as e:
                logger.error(f"Error extracting {name}: {e}")
                features[name] = {}
        
        # Add features to data
        data_dict['features'] = features
        
        # Calculate additional metrics
        data_dict['success_score'] = self._calculate_success_score(attack_data)
        data_dict['attack_hash'] = self._hash_attack(attack_data)
        
        return data_dict
    
    def _extract_payload_features(self, attack_data: AttackData) -> Dict[str, Any]:
        """Extract features from attack payload"""
        payload = attack_data.payload
        
        return {
            'length': len(payload),
            'num_tokens': len(payload.split()),
            'num_special_chars': sum(1 for c in payload if not c.isalnum() and not c.isspace()),
            'has_code': any(kw in payload.lower() for kw in ['import', 'def', 'function', '<script']),
            'has_instructions': any(kw in payload.lower() for kw in ['ignore', 'forget', 'disregard']),
            'entropy': self._calculate_entropy(payload),
            'compression_ratio': len(payload) / len(payload.encode('utf-8').hex())
        }
    
    def _extract_response_features(self, attack_data: AttackData) -> Dict[str, Any]:
        """Extract features from model response"""
        response = attack_data.response
        
        return {
            'length': len(response),
            'num_tokens': len(response.split()),
            'response_time': attack_data.response_time,
            'tokens_used': attack_data.tokens_used,
            'contains_error': any(kw in response.lower() for kw in ['error', 'cannot', 'unable']),
            'contains_warning': any(kw in response.lower() for kw in ['warning', 'caution', 'careful']),
            'sentiment': self._analyze_sentiment(response)
        }
    
    def _extract_temporal_features(self, attack_data: AttackData) -> Dict[str, Any]:
        """Extract temporal features"""
        hour = attack_data.timestamp.hour
        day_of_week = attack_data.timestamp.weekday()
        
        return {
            'hour': hour,
            'day_of_week': day_of_week,
            'is_weekend': day_of_week >= 5,
            'is_business_hours': 9 <= hour <= 17 and day_of_week < 5,
            'timestamp_unix': attack_data.timestamp.timestamp()
        }
    
    def _extract_statistical_features(self, attack_data: AttackData) -> Dict[str, Any]:
        """Extract statistical features based on historical data"""
        # Get recent attacks statistics
        recent_stats = self._get_recent_attack_stats(
            attack_data.target_model,
            attack_data.attack_type
        )
        
        return {
            'model_success_rate': recent_stats.get('success_rate', 0.0),
            'attack_type_success_rate': recent_stats.get('type_success_rate', 0.0),
            'avg_response_time': recent_stats.get('avg_response_time', 0.0),
            'detection_rate': recent_stats.get('detection_rate', 0.0)
        }
    
    def _calculate_entropy(self, text: str) -> float:
        """Calculate Shannon entropy of text"""
        if not text:
            return 0.0
            
        # Character frequency
        freq = defaultdict(int)
        for char in text:
            freq[char] += 1
            
        # Calculate entropy
        entropy = 0.0
        total = len(text)
        for count in freq.values():
            p = count / total
            if p > 0:
                entropy -= p * np.log2(p)
                
        return entropy
    
    def _analyze_sentiment(self, text: str) -> str:
        """Simple sentiment analysis (can be enhanced with ML)"""
        positive_words = {'good', 'great', 'excellent', 'success', 'yes'}
        negative_words = {'bad', 'error', 'fail', 'no', 'cannot'}
        
        text_lower = text.lower()
        positive_count = sum(1 for word in positive_words if word in text_lower)
        negative_count = sum(1 for word in negative_words if word in text_lower)
        
        if positive_count > negative_count:
            return 'positive'
        elif negative_count > positive_count:
            return 'negative'
        else:
            return 'neutral'
    
    def _calculate_success_score(self, attack_data: AttackData) -> float:
        """Calculate overall success score for the attack"""
        score = 0.0
        
        # Status contribution
        if attack_data.status == AttackStatus.SUCCESS:
            score += 1.0
        elif attack_data.status == AttackStatus.DETECTED:
            score += 0.3
        
        # Response time contribution (faster is better)
        if attack_data.response_time < 1.0:
            score += 0.2
        elif attack_data.response_time < 2.0:
            score += 0.1
            
        # Token efficiency (fewer tokens is better)
        if attack_data.tokens_used < 100:
            score += 0.2
        elif attack_data.tokens_used < 200:
            score += 0.1
            
        # Detection score (lower is better)
        score += (1 - attack_data.detection_score) * 0.3
        
        # Success indicators
        score += len(attack_data.success_indicators) * 0.1
        
        return min(score, 2.0)  # Cap at 2.0
    
    def _hash_attack(self, attack_data: AttackData) -> str:
        """Generate hash for attack deduplication"""
        hash_input = f"{attack_data.attack_type}:{attack_data.payload}:{attack_data.target_model}"
        return hashlib.sha256(hash_input.encode()).hexdigest()[:16]
    
    def _store_attack_data(self, processed_data: Dict[str, Any]):
        """Store processed attack data in database"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        
        try:
            cursor.execute('''
                INSERT OR REPLACE INTO attacks (
                    attack_id, timestamp, attack_type, target_model, provider,
                    payload, technique_params, obfuscation_level, status,
                    response, response_time, tokens_used, success_indicators,
                    detection_score, semantic_similarity, session_id, user_id,
                    campaign_id, features
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            ''', (
                processed_data['attack_id'],
                processed_data['timestamp'],
                processed_data['attack_type'],
                processed_data['target_model'],
                processed_data['provider'],
                processed_data['payload'],
                json.dumps(processed_data['technique_params']),
                processed_data['obfuscation_level'],
                processed_data['status'],
                processed_data['response'],
                processed_data['response_time'],
                processed_data['tokens_used'],
                json.dumps(processed_data['success_indicators']),
                processed_data['detection_score'],
                processed_data['semantic_similarity'],
                processed_data.get('session_id'),
                processed_data.get('user_id'),
                processed_data.get('campaign_id'),
                json.dumps(processed_data['features'])
            ))
            
            conn.commit()
            
        except Exception as e:
            logger.error(f"Error storing attack data: {e}")
            conn.rollback()
        finally:
            conn.close()
    
    def _get_recent_attack_stats(self, target_model: str, attack_type: str) -> Dict[str, float]:
        """Get statistics for recent attacks"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        
        try:
            # Get stats for last 1000 attacks
            cursor.execute('''
                SELECT 
                    AVG(CASE WHEN status = 'success' THEN 1.0 ELSE 0.0 END) as success_rate,
                    AVG(response_time) as avg_response_time,
                    AVG(CASE WHEN status = 'detected' THEN 1.0 ELSE 0.0 END) as detection_rate
                FROM (
                    SELECT status, response_time
                    FROM attacks
                    WHERE target_model = ? AND attack_type = ?
                    ORDER BY timestamp DESC
                    LIMIT 1000
                )
            ''', (target_model, attack_type))
            
            row = cursor.fetchone()
            if row:
                return {
                    'success_rate': row[0] or 0.0,
                    'avg_response_time': row[1] or 0.0,
                    'detection_rate': row[2] or 0.0
                }
                
        except Exception as e:
            logger.error(f"Error getting attack stats: {e}")
            
        finally:
            conn.close()
            
        return {}
    
    def get_training_data(self, 
                         attack_types: Optional[List[str]] = None,
                         target_models: Optional[List[str]] = None,
                         min_timestamp: Optional[datetime] = None,
                         limit: Optional[int] = None) -> pd.DataFrame:
        """
        Get processed attack data for ML training.
        
        Args:
            attack_types: Filter by attack types
            target_models: Filter by target models
            min_timestamp: Get attacks after this timestamp
            limit: Maximum number of records
            
        Returns:
            DataFrame with attack data and features
        """
        conn = sqlite3.connect(self.db_path)
        
        # Build query
        query = "SELECT * FROM attacks WHERE 1=1"
        params = []
        
        if attack_types:
            placeholders = ','.join('?' * len(attack_types))
            query += f" AND attack_type IN ({placeholders})"
            params.extend(attack_types)
            
        if target_models:
            placeholders = ','.join('?' * len(target_models))
            query += f" AND target_model IN ({placeholders})"
            params.extend(target_models)
            
        if min_timestamp:
            query += " AND timestamp > ?"
            params.append(min_timestamp.isoformat())
            
        query += " ORDER BY timestamp DESC"
        
        if limit:
            query += f" LIMIT {limit}"
            
        # Load data
        df = pd.read_sql_query(query, conn, params=params)
        conn.close()
        
        # Parse JSON columns
        for col in ['technique_params', 'success_indicators', 'features']:
            if col in df.columns:
                df[col] = df[col].apply(json.loads)
                
        return df
    
    def get_statistics(self) -> Dict[str, Any]:
        """Get pipeline statistics"""
        uptime = time.time() - self.start_time
        
        return {
            'uptime_seconds': uptime,
            'attacks_collected': self.stats['collected'],
            'attacks_processed': self.stats['processed'],
            'collection_errors': self.stats['collection_errors'],
            'processing_errors': self.stats['processing_errors'],
            'queue_size': self.data_queue.qsize(),
            'processed_queue_size': self.processed_queue.qsize(),
            'collection_rate': self.stats['collected'] / uptime if uptime > 0 else 0,
            'processing_rate': self.stats['processed'] / uptime if uptime > 0 else 0
        }
    
    def export_to_parquet(self, output_path: str, **filters):
        """Export attack data to Parquet format for efficient ML training"""
        df = self.get_training_data(**filters)
        
        # Flatten features for easier ML consumption
        if 'features' in df.columns:
            # Extract nested features
            for feature_type in ['payload_features', 'response_features', 
                               'temporal_features', 'statistical_features']:
                if feature_type in df['features'].iloc[0]:
                    feature_df = pd.json_normalize(
                        df['features'].apply(lambda x: x.get(feature_type, {}))
                    )
                    feature_df.columns = [f"{feature_type}_{col}" for col in feature_df.columns]
                    df = pd.concat([df, feature_df], axis=1)
                    
            # Drop the nested features column
            df = df.drop('features', axis=1)
            
        # Save to parquet
        df.to_parquet(output_path, index=False)
        logger.info(f"Exported {len(df)} attacks to {output_path}")
        
    def close(self):
        """Close the pipeline and clean up resources"""
        # Signal processing thread to stop
        self.processing_thread.join(timeout=5.0)
        logger.info("Attack data pipeline closed")


# Example usage for integration
def create_pipeline(config: Optional[Dict[str, Any]] = None) -> AttackDataPipeline:
    """Create and return an attack data pipeline instance"""
    default_config = {
        'storage_path': 'data/attacks',
        'batch_size': 100,
        'feature_extraction': {
            'enabled': True,
            'extractors': ['payload', 'response', 'temporal', 'statistical']
        }
    }
    
    if config:
        default_config.update(config)
        
    return AttackDataPipeline(default_config)