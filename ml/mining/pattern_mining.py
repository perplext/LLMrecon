"""
Pattern Mining for Attack Clustering in LLMrecon

This module implements advanced pattern mining techniques to discover
frequent attack patterns, association rules, and attack sequences.
"""

import numpy as np
from typing import List, Dict, Any, Tuple, Set, Optional
from dataclasses import dataclass, field
from collections import defaultdict, Counter
import itertools
from datetime import datetime
import json
import logging
import networkx as nx
from sklearn.cluster import SpectralClustering
from sklearn.metrics.pairwise import cosine_similarity
import re

logger = logging.getLogger(__name__)


@dataclass
class AttackPattern:
    """Represents a discovered attack pattern"""
    pattern_id: str
    elements: List[str]
    support: float  # Frequency in dataset
    confidence: float
    lift: float  # Statistical significance
    examples: List[str] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class SequentialPattern:
    """Represents a sequential pattern in attacks"""
    sequence: List[str]
    support: float
    examples: List[Tuple[str, int]]  # (attack_id, position)
    avg_success_rate: float
    metadata: Dict[str, Any] = field(default_factory=dict)


class PatternExtractor:
    """Extract patterns from attack text"""
    
    def __init__(self):
        # Define pattern categories
        self.pattern_types = {
            'instruction': [
                r'ignore.*(?:previous|above|prior)',
                r'disregard.*(?:instructions|guidelines)',
                r'forget.*(?:said|told)',
                r'new.*(?:instructions|directive)',
                r'override.*(?:safety|restrictions)'
            ],
            'role_play': [
                r'act as.*(?:a|an)',
                r'pretend.*(?:to be|you are)',
                r'you are.*(?:now|going to be)',
                r'roleplay.*(?:as|being)',
                r'imagine.*(?:you are|yourself as)'
            ],
            'authority': [
                r'(?:i am|i\'m).*(?:admin|developer|creator)',
                r'authorized.*(?:user|access)',
                r'security.*(?:audit|test)',
                r'official.*(?:request|directive)',
                r'system.*(?:admin|maintenance)'
            ],
            'encoding': [
                r'base64:',
                r'rot13:',
                r'\\x[0-9a-fA-F]{2}',
                r'%[0-9a-fA-F]{2}',
                r'\\u[0-9a-fA-F]{4}'
            ],
            'boundary': [
                r'\[INST\].*\[/INST\]',
                r'<\|.*\|>',
                r'###.*###',
                r'```.*```',
                r'---.*---'
            ],
            'emotional': [
                r'please.*(?:help|urgent)',
                r'(?:life|death).*(?:situation|matter)',
                r'emergency',
                r'desperate',
                r'(?:my|our).*(?:lives|family)'
            ]
        }
        
    def extract_patterns(self, text: str) -> Dict[str, List[str]]:
        """Extract all patterns from text"""
        found_patterns = defaultdict(list)
        
        for pattern_type, patterns in self.pattern_types.items():
            for pattern in patterns:
                matches = re.findall(pattern, text, re.IGNORECASE)
                if matches:
                    found_patterns[pattern_type].extend(matches)
        
        # Extract custom patterns
        found_patterns['custom'].extend(self._extract_custom_patterns(text))
        
        return dict(found_patterns)
    
    def _extract_custom_patterns(self, text: str) -> List[str]:
        """Extract custom patterns not covered by predefined categories"""
        patterns = []
        
        # Repeated characters
        repeated = re.findall(r'(.)\1{3,}', text)
        if repeated:
            patterns.append(f"repeated_chars:{len(repeated)}")
        
        # Mixed case patterns
        if any(c.islower() for c in text) and any(c.isupper() for c in text):
            case_changes = sum(1 for i in range(1, len(text)) 
                             if text[i-1].islower() and text[i].isupper())
            if case_changes > 5:
                patterns.append(f"mixed_case:{case_changes}")
        
        # Special Unicode characters
        unicode_chars = [c for c in text if ord(c) > 127]
        if unicode_chars:
            patterns.append(f"unicode:{len(unicode_chars)}")
        
        # Long words (potential obfuscation)
        words = text.split()
        long_words = [w for w in words if len(w) > 20]
        if long_words:
            patterns.append(f"long_words:{len(long_words)}")
        
        return patterns


class FrequentPatternMiner:
    """
    Mine frequent patterns using FP-Growth algorithm.
    
    Features:
    - Efficient pattern discovery
    - Support and confidence calculation
    - Association rule mining
    """
    
    def __init__(self, min_support: float = 0.01, min_confidence: float = 0.5):
        self.min_support = min_support
        self.min_confidence = min_confidence
        self.pattern_extractor = PatternExtractor()
        self.frequent_patterns: List[AttackPattern] = []
        
    def mine_patterns(self, attacks: List[Dict[str, Any]]) -> List[AttackPattern]:
        """Mine frequent patterns from attacks"""
        logger.info(f"Mining patterns from {len(attacks)} attacks")
        
        # Extract transactions (sets of patterns per attack)
        transactions = []
        attack_success = {}
        
        for attack in attacks:
            payload = attack['payload']
            patterns = self.pattern_extractor.extract_patterns(payload)
            
            # Flatten patterns into items
            items = []
            for pattern_type, pattern_list in patterns.items():
                for pattern in pattern_list:
                    items.append(f"{pattern_type}:{pattern[:30]}")  # Truncate long patterns
            
            if items:
                transactions.append(items)
                attack_success[len(transactions) - 1] = attack.get('success', False)
        
        # Build FP-tree
        fp_tree = self._build_fp_tree(transactions)
        
        # Mine frequent itemsets
        frequent_itemsets = self._mine_fp_tree(fp_tree, len(transactions))
        
        # Generate association rules
        patterns = []
        for itemset, support in frequent_itemsets.items():
            if len(itemset) >= 2:  # Only consider patterns with multiple elements
                # Calculate confidence and lift
                for i in range(1, len(itemset)):
                    for antecedent in itertools.combinations(itemset, i):
                        consequent = tuple(item for item in itemset if item not in antecedent)
                        
                        # Calculate metrics
                        antecedent_support = self._get_support(antecedent, transactions)
                        confidence = support / antecedent_support if antecedent_support > 0 else 0
                        
                        if confidence >= self.min_confidence:
                            consequent_support = self._get_support(consequent, transactions)
                            lift = confidence / consequent_support if consequent_support > 0 else 0
                            
                            # Find examples
                            examples = []
                            for idx, trans in enumerate(transactions):
                                if all(item in trans for item in itemset):
                                    if idx in attack_success and attack_success[idx]:
                                        examples.append(attacks[idx]['payload'][:100])
                            
                            pattern = AttackPattern(
                                pattern_id=f"fp_{len(patterns)}",
                                elements=list(itemset),
                                support=support,
                                confidence=confidence,
                                lift=lift,
                                examples=examples[:5],
                                metadata={
                                    'antecedent': antecedent,
                                    'consequent': consequent,
                                    'num_attacks': int(support * len(transactions))
                                }
                            )
                            patterns.append(pattern)
        
        # Sort by lift (statistical significance)
        patterns.sort(key=lambda p: p.lift, reverse=True)
        self.frequent_patterns = patterns
        
        logger.info(f"Discovered {len(patterns)} frequent patterns")
        return patterns
    
    def _build_fp_tree(self, transactions: List[List[str]]) -> Dict[str, Any]:
        """Build FP-tree from transactions"""
        # Count item frequencies
        item_counts = Counter()
        for trans in transactions:
            item_counts.update(trans)
        
        # Filter by minimum support
        min_count = self.min_support * len(transactions)
        frequent_items = {item: count for item, count in item_counts.items() 
                         if count >= min_count}
        
        # Build tree
        tree = {'root': {'count': 0, 'children': {}}}
        
        for trans in transactions:
            # Filter and sort by frequency
            filtered_trans = [item for item in trans if item in frequent_items]
            filtered_trans.sort(key=lambda x: frequent_items[x], reverse=True)
            
            # Insert into tree
            current = tree['root']
            for item in filtered_trans:
                if item not in current['children']:
                    current['children'][item] = {'count': 0, 'children': {}}
                current = current['children'][item]
                current['count'] += 1
        
        return tree
    
    def _mine_fp_tree(self, tree: Dict[str, Any], num_transactions: int) -> Dict[Tuple[str], float]:
        """Mine frequent itemsets from FP-tree"""
        frequent_itemsets = {}
        
        def mine_subtree(node, prefix, min_count):
            for item, child in node['children'].items():
                if child['count'] >= min_count:
                    new_prefix = prefix + (item,)
                    support = child['count'] / num_transactions
                    frequent_itemsets[new_prefix] = support
                    mine_subtree(child, new_prefix, min_count)
        
        min_count = self.min_support * num_transactions
        mine_subtree(tree['root'], (), min_count)
        
        return frequent_itemsets
    
    def _get_support(self, itemset: Tuple[str], transactions: List[List[str]]) -> float:
        """Calculate support for an itemset"""
        count = sum(1 for trans in transactions if all(item in trans for item in itemset))
        return count / len(transactions) if transactions else 0


class SequentialPatternMiner:
    """
    Mine sequential patterns in attack sequences.
    
    Features:
    - Temporal pattern discovery
    - Multi-step attack detection
    - Success rate correlation
    """
    
    def __init__(self, min_support: float = 0.05, max_gap: int = 5):
        self.min_support = min_support
        self.max_gap = max_gap
        self.sequential_patterns: List[SequentialPattern] = []
        
    def mine_sequences(self, attack_sessions: List[List[Dict[str, Any]]]) -> List[SequentialPattern]:
        """
        Mine sequential patterns from attack sessions.
        
        Args:
            attack_sessions: List of attack sequences (each session is a list of attacks)
        """
        logger.info(f"Mining sequential patterns from {len(attack_sessions)} sessions")
        
        # Convert attacks to symbolic sequences
        symbol_sequences = []
        success_rates = defaultdict(list)
        
        for session in attack_sessions:
            sequence = []
            for attack in session:
                # Create symbol from attack characteristics
                symbol = self._attack_to_symbol(attack)
                sequence.append(symbol)
                success_rates[symbol].append(1 if attack.get('success', False) else 0)
            
            if sequence:
                symbol_sequences.append(sequence)
        
        # Calculate average success rates
        avg_success_rates = {
            symbol: np.mean(rates) for symbol, rates in success_rates.items()
        }
        
        # Mine sequential patterns using PrefixSpan
        patterns = self._prefix_span(symbol_sequences)
        
        # Create SequentialPattern objects
        sequential_patterns = []
        for pattern, support in patterns.items():
            if len(pattern) >= 2:  # Only multi-step patterns
                examples = self._find_pattern_occurrences(pattern, symbol_sequences)
                
                # Calculate average success rate for pattern
                pattern_success_rates = [avg_success_rates.get(symbol, 0) for symbol in pattern]
                avg_pattern_success = np.mean(pattern_success_rates)
                
                seq_pattern = SequentialPattern(
                    sequence=list(pattern),
                    support=support,
                    examples=examples[:10],
                    avg_success_rate=avg_pattern_success,
                    metadata={
                        'length': len(pattern),
                        'unique_symbols': len(set(pattern))
                    }
                )
                sequential_patterns.append(seq_pattern)
        
        # Sort by support
        sequential_patterns.sort(key=lambda p: p.support, reverse=True)
        self.sequential_patterns = sequential_patterns
        
        logger.info(f"Discovered {len(sequential_patterns)} sequential patterns")
        return sequential_patterns
    
    def _attack_to_symbol(self, attack: Dict[str, Any]) -> str:
        """Convert attack to symbolic representation"""
        # Extract key features
        attack_type = attack.get('attack_type', 'unknown')
        success = 'S' if attack.get('success', False) else 'F'
        
        # Simplified pattern detection
        payload = attack.get('payload', '').lower()
        if 'ignore' in payload or 'disregard' in payload:
            technique = 'IGN'
        elif 'act as' in payload or 'pretend' in payload:
            technique = 'ROLE'
        elif 'base64' in payload or '\\x' in payload:
            technique = 'ENC'
        else:
            technique = 'STD'
        
        return f"{attack_type[:3].upper()}_{technique}_{success}"
    
    def _prefix_span(self, sequences: List[List[str]]) -> Dict[Tuple[str], float]:
        """PrefixSpan algorithm for sequential pattern mining"""
        min_count = self.min_support * len(sequences)
        patterns = {}
        
        # Find frequent 1-sequences
        item_counts = Counter()
        for seq in sequences:
            item_counts.update(set(seq))
        
        frequent_items = {item: count for item, count in item_counts.items() 
                         if count >= min_count}
        
        # Recursive mining
        def mine_prefix(prefix, projected_db):
            for item in frequent_items:
                # Count occurrences
                count = 0
                new_projected_db = []
                
                for seq in projected_db:
                    # Find first occurrence after prefix
                    pos = self._find_first_occurrence(seq, item)
                    if pos is not None:
                        count += 1
                        if pos + 1 < len(seq):
                            new_projected_db.append(seq[pos + 1:])
                
                if count >= min_count:
                    new_pattern = prefix + (item,)
                    patterns[new_pattern] = count / len(sequences)
                    
                    if new_projected_db:
                        mine_prefix(new_pattern, new_projected_db)
        
        # Start mining with each frequent item
        for item in frequent_items:
            projected_db = []
            for seq in sequences:
                pos = self._find_first_occurrence(seq, item)
                if pos is not None and pos + 1 < len(seq):
                    projected_db.append(seq[pos + 1:])
            
            patterns[(item,)] = frequent_items[item] / len(sequences)
            if projected_db:
                mine_prefix((item,), projected_db)
        
        return patterns
    
    def _find_first_occurrence(self, sequence: List[str], item: str) -> Optional[int]:
        """Find first occurrence of item in sequence"""
        try:
            return sequence.index(item)
        except ValueError:
            return None
    
    def _find_pattern_occurrences(self, 
                                 pattern: Tuple[str], 
                                 sequences: List[List[str]]) -> List[Tuple[str, int]]:
        """Find all occurrences of a pattern in sequences"""
        occurrences = []
        
        for seq_idx, seq in enumerate(sequences):
            # Find pattern in sequence
            for start_pos in range(len(seq) - len(pattern) + 1):
                # Check if pattern matches with gaps allowed
                match = True
                pos = start_pos
                
                for pattern_item in pattern:
                    found = False
                    for gap in range(min(self.max_gap + 1, len(seq) - pos)):
                        if pos + gap < len(seq) and seq[pos + gap] == pattern_item:
                            pos = pos + gap + 1
                            found = True
                            break
                    
                    if not found:
                        match = False
                        break
                
                if match:
                    occurrences.append((f"session_{seq_idx}", start_pos))
        
        return occurrences


class AttackGraphMiner:
    """
    Build and analyze attack dependency graphs.
    
    Features:
    - Attack relationship discovery
    - Success path analysis
    - Vulnerability chain detection
    """
    
    def __init__(self):
        self.graph = nx.DiGraph()
        self.pattern_extractor = PatternExtractor()
        
    def build_attack_graph(self, attacks: List[Dict[str, Any]]) -> nx.DiGraph:
        """Build graph of attack relationships"""
        logger.info(f"Building attack graph from {len(attacks)} attacks")
        
        # Add nodes for each attack
        for idx, attack in enumerate(attacks):
            patterns = self.pattern_extractor.extract_patterns(attack['payload'])
            
            self.graph.add_node(
                idx,
                payload=attack['payload'][:100],
                success=attack.get('success', False),
                patterns=patterns,
                attack_type=attack.get('attack_type', 'unknown')
            )
        
        # Add edges based on pattern similarity
        for i in range(len(attacks)):
            for j in range(i + 1, len(attacks)):
                similarity = self._calculate_pattern_similarity(
                    self.graph.nodes[i]['patterns'],
                    self.graph.nodes[j]['patterns']
                )
                
                if similarity > 0.5:  # Threshold for connection
                    self.graph.add_edge(i, j, weight=similarity)
                    self.graph.add_edge(j, i, weight=similarity)
        
        logger.info(f"Built graph with {self.graph.number_of_nodes()} nodes and {self.graph.number_of_edges()} edges")
        return self.graph
    
    def _calculate_pattern_similarity(self, patterns1: Dict, patterns2: Dict) -> float:
        """Calculate similarity between two pattern sets"""
        # Create feature vectors
        all_patterns = set()
        for p in [patterns1, patterns2]:
            for pattern_type, pattern_list in p.items():
                all_patterns.update(f"{pattern_type}:{p}" for p in pattern_list)
        
        if not all_patterns:
            return 0.0
        
        # Binary vectors
        vec1 = np.array([1 if p in str(patterns1) else 0 for p in all_patterns])
        vec2 = np.array([1 if p in str(patterns2) else 0 for p in all_patterns])
        
        # Jaccard similarity
        intersection = np.sum(vec1 * vec2)
        union = np.sum(np.maximum(vec1, vec2))
        
        return intersection / union if union > 0 else 0.0
    
    def find_attack_communities(self) -> List[Set[int]]:
        """Find communities of related attacks"""
        # Use spectral clustering on graph
        if self.graph.number_of_nodes() < 2:
            return []
        
        # Create adjacency matrix
        adj_matrix = nx.adjacency_matrix(self.graph).todense()
        
        # Spectral clustering
        n_clusters = min(10, self.graph.number_of_nodes() // 5)
        clustering = SpectralClustering(
            n_clusters=n_clusters,
            affinity='precomputed',
            random_state=42
        )
        
        labels = clustering.fit_predict(adj_matrix)
        
        # Group nodes by cluster
        communities = defaultdict(set)
        for node, label in enumerate(labels):
            communities[label].add(node)
        
        return list(communities.values())
    
    def analyze_success_paths(self) -> List[List[int]]:
        """Find paths that lead to successful attacks"""
        successful_nodes = [
            n for n in self.graph.nodes()
            if self.graph.nodes[n].get('success', False)
        ]
        
        paths = []
        
        # Find paths to successful nodes
        for target in successful_nodes:
            # Use BFS to find paths
            for source in self.graph.nodes():
                if source != target:
                    try:
                        path = nx.shortest_path(self.graph, source, target, weight='weight')
                        if len(path) > 1:
                            paths.append(path)
                    except nx.NetworkXNoPath:
                        continue
        
        # Filter for interesting paths
        interesting_paths = []
        for path in paths:
            # Check if path shows progression (e.g., failed -> successful)
            if not self.graph.nodes[path[0]].get('success', False) and \
               self.graph.nodes[path[-1]].get('success', False):
                interesting_paths.append(path)
        
        return interesting_paths


class IntegratedPatternMining:
    """
    Integrated pattern mining system combining all techniques.
    
    Features:
    - Unified pattern discovery
    - Cross-technique validation
    - Comprehensive reporting
    """
    
    def __init__(self):
        self.frequent_miner = FrequentPatternMiner()
        self.sequential_miner = SequentialPatternMiner()
        self.graph_miner = AttackGraphMiner()
        self.discovered_insights = {}
        
    def analyze_attack_corpus(self, 
                            attacks: List[Dict[str, Any]],
                            sessions: Optional[List[List[Dict[str, Any]]]] = None) -> Dict[str, Any]:
        """
        Comprehensive pattern analysis of attack corpus.
        
        Args:
            attacks: List of all attacks
            sessions: Optional grouped attack sessions
        """
        logger.info("Starting comprehensive pattern mining analysis")
        
        # Frequent pattern mining
        frequent_patterns = self.frequent_miner.mine_patterns(attacks)
        
        # Sequential pattern mining
        sequential_patterns = []
        if sessions:
            sequential_patterns = self.sequential_miner.mine_sequences(sessions)
        
        # Graph-based analysis
        attack_graph = self.graph_miner.build_attack_graph(attacks)
        communities = self.graph_miner.find_attack_communities()
        success_paths = self.graph_miner.analyze_success_paths()
        
        # Combine insights
        self.discovered_insights = {
            'frequent_patterns': {
                'count': len(frequent_patterns),
                'top_patterns': [
                    {
                        'elements': p.elements,
                        'support': p.support,
                        'confidence': p.confidence,
                        'lift': p.lift
                    }
                    for p in frequent_patterns[:10]
                ]
            },
            'sequential_patterns': {
                'count': len(sequential_patterns),
                'top_sequences': [
                    {
                        'sequence': p.sequence,
                        'support': p.support,
                        'avg_success_rate': p.avg_success_rate
                    }
                    for p in sequential_patterns[:10]
                ]
            },
            'graph_analysis': {
                'nodes': attack_graph.number_of_nodes(),
                'edges': attack_graph.number_of_edges(),
                'communities': len(communities),
                'largest_community': max(len(c) for c in communities) if communities else 0,
                'success_paths': len(success_paths)
            },
            'pattern_categories': self._categorize_patterns(frequent_patterns),
            'vulnerability_chains': self._extract_vulnerability_chains(success_paths, attacks)
        }
        
        return self.discovered_insights
    
    def _categorize_patterns(self, patterns: List[AttackPattern]) -> Dict[str, int]:
        """Categorize patterns by type"""
        categories = defaultdict(int)
        
        for pattern in patterns:
            for element in pattern.elements:
                if ':' in element:
                    category = element.split(':')[0]
                    categories[category] += 1
        
        return dict(categories)
    
    def _extract_vulnerability_chains(self, 
                                    paths: List[List[int]], 
                                    attacks: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """Extract vulnerability chains from successful paths"""
        chains = []
        
        for path in paths[:10]:  # Top 10 chains
            chain = {
                'length': len(path),
                'steps': []
            }
            
            for node in path:
                if node < len(attacks):
                    attack = attacks[node]
                    chain['steps'].append({
                        'attack_type': attack.get('attack_type', 'unknown'),
                        'success': attack.get('success', False),
                        'payload_preview': attack['payload'][:50] + '...'
                    })
            
            chains.append(chain)
        
        return chains
    
    def generate_report(self) -> str:
        """Generate human-readable report of discoveries"""
        report = []
        report.append("# Pattern Mining Analysis Report\n")
        report.append(f"Generated: {datetime.now().isoformat()}\n")
        
        # Executive Summary
        report.append("## Executive Summary\n")
        insights = self.discovered_insights
        
        report.append(f"- Discovered {insights['frequent_patterns']['count']} frequent attack patterns")
        report.append(f"- Found {insights['sequential_patterns']['count']} sequential attack patterns")
        report.append(f"- Identified {insights['graph_analysis']['communities']} attack communities")
        report.append(f"- Detected {insights['graph_analysis']['success_paths']} successful attack paths\n")
        
        # Top Patterns
        report.append("## Top Frequent Patterns\n")
        for i, pattern in enumerate(insights['frequent_patterns']['top_patterns'][:5]):
            report.append(f"{i+1}. Pattern: {' + '.join(pattern['elements'][:3])}")
            report.append(f"   - Support: {pattern['support']:.2%}")
            report.append(f"   - Confidence: {pattern['confidence']:.2%}")
            report.append(f"   - Lift: {pattern['lift']:.2f}\n")
        
        # Sequential Patterns
        if insights['sequential_patterns']['top_sequences']:
            report.append("## Top Sequential Patterns\n")
            for i, seq in enumerate(insights['sequential_patterns']['top_sequences'][:5]):
                report.append(f"{i+1}. Sequence: {' → '.join(seq['sequence'])}")
                report.append(f"   - Support: {seq['support']:.2%}")
                report.append(f"   - Success Rate: {seq['avg_success_rate']:.2%}\n")
        
        # Pattern Categories
        report.append("## Pattern Distribution\n")
        for category, count in insights['pattern_categories'].items():
            report.append(f"- {category}: {count} patterns")
        
        return '\n'.join(report)
    
    def export_results(self, filepath: str):
        """Export all results to file"""
        export_data = {
            'timestamp': datetime.now().isoformat(),
            'insights': self.discovered_insights,
            'frequent_patterns': [
                {
                    'pattern_id': p.pattern_id,
                    'elements': p.elements,
                    'support': p.support,
                    'confidence': p.confidence,
                    'lift': p.lift,
                    'examples': p.examples
                }
                for p in self.frequent_miner.frequent_patterns
            ],
            'sequential_patterns': [
                {
                    'sequence': p.sequence,
                    'support': p.support,
                    'avg_success_rate': p.avg_success_rate,
                    'examples': p.examples
                }
                for p in self.sequential_miner.sequential_patterns
            ]
        }
        
        with open(filepath, 'w') as f:
            json.dump(export_data, f, indent=2)
        
        logger.info(f"Exported pattern mining results to {filepath}")