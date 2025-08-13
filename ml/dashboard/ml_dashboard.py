"""
ML Performance Dashboard for LLMrecon

This module provides a comprehensive dashboard for monitoring and analyzing
ML model performance, training progress, and attack effectiveness.
"""

import streamlit as st
import plotly.graph_objs as go
import plotly.express as px
from plotly.subplots import make_subplots
import pandas as pd
import numpy as np
from datetime import datetime, timedelta
import json
import sqlite3
from typing import Dict, List, Any, Optional, Tuple
import logging
from pathlib import Path

logger = logging.getLogger(__name__)


class MLDashboard:
    """
    Main dashboard class for ML performance monitoring.
    
    Features:
    - Real-time model performance tracking
    - Training progress visualization
    - Attack success analysis
    - Model comparison tools
    """
    
    def __init__(self, data_path: str = "ml/data"):
        self.data_path = Path(data_path)
        self.db_path = self.data_path / "ml_metrics.db"
        self._init_database()
    
    def _init_database(self):
        """Initialize metrics database"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        
        # Model metrics table
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS model_metrics (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                model_id TEXT NOT NULL,
                model_type TEXT NOT NULL,
                timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
                metric_name TEXT NOT NULL,
                metric_value REAL NOT NULL,
                metadata TEXT
            )
        ''')
        
        # Training progress table
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS training_progress (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                model_id TEXT NOT NULL,
                epoch INTEGER NOT NULL,
                timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
                loss REAL,
                accuracy REAL,
                val_loss REAL,
                val_accuracy REAL,
                learning_rate REAL,
                metadata TEXT
            )
        ''')
        
        # Attack results table
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS attack_results (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                attack_id TEXT NOT NULL,
                model_id TEXT NOT NULL,
                attack_type TEXT NOT NULL,
                timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
                success BOOLEAN,
                confidence REAL,
                response_time REAL,
                metadata TEXT
            )
        ''')
        
        conn.commit()
        conn.close()
    
    def run(self):
        """Run the Streamlit dashboard"""
        st.set_page_config(
            page_title="LLMrecon ML Dashboard",
            page_icon="🤖",
            layout="wide"
        )
        
        st.title("🤖 LLMrecon ML Performance Dashboard")
        
        # Sidebar navigation
        page = st.sidebar.selectbox(
            "Select Page",
            ["Overview", "Model Performance", "Training Monitor", 
             "Attack Analytics", "Model Comparison", "System Health"]
        )
        
        if page == "Overview":
            self._show_overview()
        elif page == "Model Performance":
            self._show_model_performance()
        elif page == "Training Monitor":
            self._show_training_monitor()
        elif page == "Attack Analytics":
            self._show_attack_analytics()
        elif page == "Model Comparison":
            self._show_model_comparison()
        elif page == "System Health":
            self._show_system_health()
    
    def _show_overview(self):
        """Show overview page"""
        st.header("System Overview")
        
        # Key metrics
        col1, col2, col3, col4 = st.columns(4)
        
        with col1:
            total_models = self._get_total_models()
            st.metric("Total Models", total_models, "+2")
        
        with col2:
            active_training = self._get_active_training()
            st.metric("Active Training", active_training, "+1")
        
        with col3:
            avg_success_rate = self._get_avg_success_rate()
            st.metric("Avg Success Rate", f"{avg_success_rate:.1%}", "+2.3%")
        
        with col4:
            total_attacks = self._get_total_attacks()
            st.metric("Total Attacks", f"{total_attacks:,}", "+523")
        
        # Recent activity
        st.subheader("Recent Activity")
        
        col1, col2 = st.columns(2)
        
        with col1:
            st.write("**Recent Model Updates**")
            recent_models = self._get_recent_models()
            for model in recent_models:
                st.write(f"- {model['model_id']} ({model['model_type']}) - {model['timestamp']}")
        
        with col2:
            st.write("**Recent Successful Attacks**")
            recent_attacks = self._get_recent_successful_attacks()
            for attack in recent_attacks:
                st.write(f"- {attack['attack_type']} on {attack['model_id']} - {attack['confidence']:.2f}")
        
        # Performance trends
        st.subheader("Performance Trends")
        
        trends_data = self._get_performance_trends()
        fig = make_subplots(
            rows=2, cols=2,
            subplot_titles=('Success Rate Trend', 'Response Time Trend', 
                          'Model Accuracy Trend', 'Attack Volume Trend')
        )
        
        # Success rate
        fig.add_trace(
            go.Scatter(x=trends_data['dates'], y=trends_data['success_rates'],
                      mode='lines+markers', name='Success Rate'),
            row=1, col=1
        )
        
        # Response time
        fig.add_trace(
            go.Scatter(x=trends_data['dates'], y=trends_data['response_times'],
                      mode='lines+markers', name='Response Time', line=dict(color='orange')),
            row=1, col=2
        )
        
        # Model accuracy
        fig.add_trace(
            go.Scatter(x=trends_data['dates'], y=trends_data['accuracies'],
                      mode='lines+markers', name='Accuracy', line=dict(color='green')),
            row=2, col=1
        )
        
        # Attack volume
        fig.add_trace(
            go.Bar(x=trends_data['dates'], y=trends_data['attack_counts'],
                   name='Attacks', marker=dict(color='red')),
            row=2, col=2
        )
        
        fig.update_layout(height=800, showlegend=False)
        st.plotly_chart(fig, use_container_width=True)
    
    def _show_model_performance(self):
        """Show model performance page"""
        st.header("Model Performance Analysis")
        
        # Model selector
        models = self._get_all_models()
        selected_model = st.selectbox("Select Model", models)
        
        if selected_model:
            # Model info
            col1, col2, col3 = st.columns(3)
            
            model_info = self._get_model_info(selected_model)
            
            with col1:
                st.metric("Model Type", model_info['type'])
                st.metric("Total Predictions", model_info['predictions'])
            
            with col2:
                st.metric("Success Rate", f"{model_info['success_rate']:.1%}")
                st.metric("Avg Confidence", f"{model_info['avg_confidence']:.2f}")
            
            with col3:
                st.metric("Avg Response Time", f"{model_info['avg_response']:.2f}s")
                st.metric("Last Updated", model_info['last_updated'])
            
            # Performance charts
            st.subheader("Performance Metrics")
            
            metrics_data = self._get_model_metrics(selected_model)
            
            # Metrics over time
            fig = make_subplots(
                rows=2, cols=2,
                subplot_titles=('Loss', 'Accuracy', 'Success Rate', 'Response Time')
            )
            
            # Loss
            fig.add_trace(
                go.Scatter(x=metrics_data['timestamps'], y=metrics_data['loss'],
                          mode='lines', name='Loss'),
                row=1, col=1
            )
            
            # Accuracy
            fig.add_trace(
                go.Scatter(x=metrics_data['timestamps'], y=metrics_data['accuracy'],
                          mode='lines', name='Accuracy', line=dict(color='green')),
                row=1, col=2
            )
            
            # Success rate
            fig.add_trace(
                go.Scatter(x=metrics_data['timestamps'], y=metrics_data['success_rate'],
                          mode='lines', name='Success Rate', line=dict(color='blue')),
                row=2, col=1
            )
            
            # Response time
            fig.add_trace(
                go.Scatter(x=metrics_data['timestamps'], y=metrics_data['response_time'],
                          mode='lines', name='Response Time', line=dict(color='orange')),
                row=2, col=2
            )
            
            fig.update_layout(height=600, showlegend=False)
            st.plotly_chart(fig, use_container_width=True)
            
            # Feature importance (if available)
            if 'feature_importance' in model_info:
                st.subheader("Feature Importance")
                
                importance_df = pd.DataFrame(model_info['feature_importance'])
                fig = px.bar(importance_df, x='importance', y='feature', 
                           orientation='h', title='Top Features')
                st.plotly_chart(fig, use_container_width=True)
    
    def _show_training_monitor(self):
        """Show training monitor page"""
        st.header("Training Monitor")
        
        # Active training sessions
        active_sessions = self._get_active_training_sessions()
        
        if active_sessions:
            st.subheader("Active Training Sessions")
            
            for session in active_sessions:
                with st.expander(f"{session['model_id']} - Epoch {session['current_epoch']}"):
                    col1, col2 = st.columns(2)
                    
                    with col1:
                        st.metric("Progress", f"{session['progress']:.0%}")
                        st.metric("Current Loss", f"{session['current_loss']:.4f}")
                        st.metric("Current Accuracy", f"{session['current_accuracy']:.2%}")
                    
                    with col2:
                        st.metric("Learning Rate", f"{session['learning_rate']:.6f}")
                        st.metric("Time Elapsed", session['time_elapsed'])
                        st.metric("ETA", session['eta'])
                    
                    # Progress bar
                    st.progress(session['progress'])
                    
                    # Live loss plot
                    loss_data = session['loss_history']
                    fig = go.Figure()
                    fig.add_trace(go.Scatter(y=loss_data, mode='lines', name='Training Loss'))
                    fig.update_layout(height=300, title='Training Loss')
                    st.plotly_chart(fig, use_container_width=True)
        else:
            st.info("No active training sessions")
        
        # Training history
        st.subheader("Training History")
        
        history_df = self._get_training_history()
        
        if not history_df.empty:
            # Filter options
            col1, col2 = st.columns(2)
            with col1:
                selected_models = st.multiselect(
                    "Select Models",
                    history_df['model_id'].unique(),
                    default=history_df['model_id'].unique()[:3]
                )
            
            with col2:
                date_range = st.date_input(
                    "Date Range",
                    value=(datetime.now() - timedelta(days=7), datetime.now()),
                    max_value=datetime.now()
                )
            
            # Filter data
            filtered_df = history_df[
                (history_df['model_id'].isin(selected_models)) &
                (history_df['timestamp'] >= pd.Timestamp(date_range[0])) &
                (history_df['timestamp'] <= pd.Timestamp(date_range[1]))
            ]
            
            # Loss comparison
            fig = px.line(filtered_df, x='epoch', y='loss', color='model_id',
                         title='Training Loss Comparison')
            st.plotly_chart(fig, use_container_width=True)
            
            # Accuracy comparison
            fig = px.line(filtered_df, x='epoch', y='accuracy', color='model_id',
                         title='Training Accuracy Comparison')
            st.plotly_chart(fig, use_container_width=True)
    
    def _show_attack_analytics(self):
        """Show attack analytics page"""
        st.header("Attack Analytics")
        
        # Attack statistics
        col1, col2, col3, col4 = st.columns(4)
        
        stats = self._get_attack_statistics()
        
        with col1:
            st.metric("Total Attacks", f"{stats['total']:,}")
            st.metric("Unique Types", stats['unique_types'])
        
        with col2:
            st.metric("Success Rate", f"{stats['success_rate']:.1%}")
            st.metric("Avg Confidence", f"{stats['avg_confidence']:.2f}")
        
        with col3:
            st.metric("Today's Attacks", stats['today_count'])
            st.metric("This Week", stats['week_count'])
        
        with col4:
            st.metric("Best Performer", stats['best_attack_type'])
            st.metric("Most Targeted", stats['most_targeted_model'])
        
        # Attack type breakdown
        st.subheader("Attack Type Analysis")
        
        attack_breakdown = self._get_attack_breakdown()
        
        col1, col2 = st.columns(2)
        
        with col1:
            # Pie chart of attack types
            fig = px.pie(attack_breakdown, values='count', names='attack_type',
                        title='Attack Type Distribution')
            st.plotly_chart(fig, use_container_width=True)
        
        with col2:
            # Success rate by type
            fig = px.bar(attack_breakdown, x='attack_type', y='success_rate',
                        title='Success Rate by Attack Type', color='success_rate',
                        color_continuous_scale='RdYlGn')
            fig.update_layout(xaxis_tickangle=-45)
            st.plotly_chart(fig, use_container_width=True)
        
        # Time series analysis
        st.subheader("Attack Patterns Over Time")
        
        time_series = self._get_attack_time_series()
        
        # Hourly pattern
        fig = px.line(time_series['hourly'], x='hour', y='count',
                     title='Attacks by Hour of Day')
        st.plotly_chart(fig, use_container_width=True)
        
        # Success rate heatmap
        st.subheader("Success Rate Heatmap")
        
        heatmap_data = self._get_success_heatmap()
        
        fig = px.imshow(heatmap_data, 
                       labels=dict(x="Target Model", y="Attack Type", color="Success Rate"),
                       title="Attack Success Rate by Type and Target",
                       color_continuous_scale='RdYlGn')
        st.plotly_chart(fig, use_container_width=True)
        
        # Recent attacks table
        st.subheader("Recent Attack Details")
        
        recent_attacks = self._get_recent_attacks_details()
        st.dataframe(recent_attacks, use_container_width=True)
    
    def _show_model_comparison(self):
        """Show model comparison page"""
        st.header("Model Comparison")
        
        # Model selection
        all_models = self._get_all_models()
        selected_models = st.multiselect(
            "Select Models to Compare",
            all_models,
            default=all_models[:3] if len(all_models) >= 3 else all_models
        )
        
        if len(selected_models) >= 2:
            # Comparison metrics
            comparison_data = self._get_model_comparison_data(selected_models)
            
            # Radar chart
            st.subheader("Performance Radar Chart")
            
            categories = ['Success Rate', 'Accuracy', 'Speed', 'Robustness', 'Efficiency']
            
            fig = go.Figure()
            
            for model in selected_models:
                values = comparison_data[model]['radar_values']
                fig.add_trace(go.Scatterpolar(
                    r=values,
                    theta=categories,
                    fill='toself',
                    name=model
                ))
            
            fig.update_layout(
                polar=dict(
                    radialaxis=dict(
                        visible=True,
                        range=[0, 1]
                    )),
                showlegend=True
            )
            
            st.plotly_chart(fig, use_container_width=True)
            
            # Side-by-side metrics
            st.subheader("Detailed Comparison")
            
            metrics_df = pd.DataFrame(comparison_data).T
            st.dataframe(metrics_df, use_container_width=True)
            
            # Head-to-head performance
            st.subheader("Head-to-Head Performance")
            
            if len(selected_models) == 2:
                model1, model2 = selected_models
                h2h_data = self._get_head_to_head(model1, model2)
                
                col1, col2, col3 = st.columns(3)
                
                with col1:
                    st.metric(f"{model1} Wins", h2h_data['model1_wins'])
                
                with col2:
                    st.metric("Draws", h2h_data['draws'])
                
                with col3:
                    st.metric(f"{model2} Wins", h2h_data['model2_wins'])
                
                # Common attacks comparison
                fig = go.Figure()
                fig.add_trace(go.Bar(name=model1, x=h2h_data['attack_types'], 
                                   y=h2h_data['model1_success']))
                fig.add_trace(go.Bar(name=model2, x=h2h_data['attack_types'], 
                                   y=h2h_data['model2_success']))
                
                fig.update_layout(barmode='group', title='Success Rate on Common Attacks')
                st.plotly_chart(fig, use_container_width=True)
        else:
            st.warning("Please select at least 2 models for comparison")
    
    def _show_system_health(self):
        """Show system health page"""
        st.header("System Health Monitor")
        
        # System metrics
        col1, col2, col3, col4 = st.columns(4)
        
        health_metrics = self._get_system_health_metrics()
        
        with col1:
            status_color = "🟢" if health_metrics['cpu_usage'] < 80 else "🔴"
            st.metric(f"{status_color} CPU Usage", f"{health_metrics['cpu_usage']:.1f}%")
            st.metric("Memory Usage", f"{health_metrics['memory_usage']:.1f}%")
        
        with col2:
            st.metric("GPU Usage", f"{health_metrics['gpu_usage']:.1f}%")
            st.metric("GPU Memory", f"{health_metrics['gpu_memory']:.1f}%")
        
        with col3:
            st.metric("Disk Usage", f"{health_metrics['disk_usage']:.1f}%")
            st.metric("Network I/O", f"{health_metrics['network_io']:.1f} MB/s")
        
        with col4:
            st.metric("Active Jobs", health_metrics['active_jobs'])
            st.metric("Queue Length", health_metrics['queue_length'])
        
        # Resource usage over time
        st.subheader("Resource Usage Trends")
        
        resource_history = self._get_resource_history()
        
        fig = make_subplots(
            rows=2, cols=2,
            subplot_titles=('CPU Usage', 'Memory Usage', 'GPU Usage', 'Disk I/O')
        )
        
        # CPU
        fig.add_trace(
            go.Scatter(x=resource_history['timestamps'], y=resource_history['cpu'],
                      mode='lines', name='CPU'),
            row=1, col=1
        )
        
        # Memory
        fig.add_trace(
            go.Scatter(x=resource_history['timestamps'], y=resource_history['memory'],
                      mode='lines', name='Memory', line=dict(color='orange')),
            row=1, col=2
        )
        
        # GPU
        fig.add_trace(
            go.Scatter(x=resource_history['timestamps'], y=resource_history['gpu'],
                      mode='lines', name='GPU', line=dict(color='green')),
            row=2, col=1
        )
        
        # Disk I/O
        fig.add_trace(
            go.Scatter(x=resource_history['timestamps'], y=resource_history['disk_io'],
                      mode='lines', name='Disk I/O', line=dict(color='red')),
            row=2, col=2
        )
        
        fig.update_layout(height=600, showlegend=False)
        st.plotly_chart(fig, use_container_width=True)
        
        # Error logs
        st.subheader("Recent Errors and Warnings")
        
        errors = self._get_recent_errors()
        
        if errors:
            for error in errors:
                severity = "🔴" if error['level'] == 'ERROR' else "🟡"
                with st.expander(f"{severity} {error['timestamp']} - {error['component']}"):
                    st.code(error['message'])
                    if error['stacktrace']:
                        st.code(error['stacktrace'])
        else:
            st.success("No recent errors")
        
        # Model health check
        st.subheader("Model Health Status")
        
        model_health = self._get_model_health_status()
        
        health_df = pd.DataFrame(model_health)
        
        # Color code based on status
        def color_status(val):
            if val == 'Healthy':
                return 'background-color: #90EE90'
            elif val == 'Warning':
                return 'background-color: #FFD700'
            else:
                return 'background-color: #FFB6C1'
        
        styled_df = health_df.style.applymap(color_status, subset=['Status'])
        st.dataframe(styled_df, use_container_width=True)
    
    # Data retrieval methods (implement based on your data storage)
    def _get_total_models(self) -> int:
        """Get total number of models"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        cursor.execute("SELECT COUNT(DISTINCT model_id) FROM model_metrics")
        count = cursor.fetchone()[0]
        conn.close()
        return count
    
    def _get_active_training(self) -> int:
        """Get number of active training sessions"""
        # Simulate for demo
        return np.random.randint(0, 5)
    
    def _get_avg_success_rate(self) -> float:
        """Get average success rate"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        cursor.execute("SELECT AVG(success) FROM attack_results")
        rate = cursor.fetchone()[0] or 0.0
        conn.close()
        return rate
    
    def _get_total_attacks(self) -> int:
        """Get total number of attacks"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        cursor.execute("SELECT COUNT(*) FROM attack_results")
        count = cursor.fetchone()[0]
        conn.close()
        return count
    
    def _get_recent_models(self) -> List[Dict[str, Any]]:
        """Get recent model updates"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        cursor.execute("""
            SELECT DISTINCT model_id, model_type, MAX(timestamp) as timestamp
            FROM model_metrics
            GROUP BY model_id, model_type
            ORDER BY timestamp DESC
            LIMIT 5
        """)
        models = []
        for row in cursor.fetchall():
            models.append({
                'model_id': row[0],
                'model_type': row[1],
                'timestamp': row[2]
            })
        conn.close()
        return models
    
    def _get_recent_successful_attacks(self) -> List[Dict[str, Any]]:
        """Get recent successful attacks"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        cursor.execute("""
            SELECT attack_type, model_id, confidence
            FROM attack_results
            WHERE success = 1
            ORDER BY timestamp DESC
            LIMIT 5
        """)
        attacks = []
        for row in cursor.fetchall():
            attacks.append({
                'attack_type': row[0],
                'model_id': row[1],
                'confidence': row[2]
            })
        conn.close()
        return attacks
    
    def _get_performance_trends(self) -> Dict[str, List[Any]]:
        """Get performance trends data"""
        # Simulate trend data for demo
        dates = pd.date_range(end=datetime.now(), periods=30, freq='D')
        
        return {
            'dates': dates.tolist(),
            'success_rates': np.random.uniform(0.6, 0.9, 30).tolist(),
            'response_times': np.random.uniform(0.5, 2.0, 30).tolist(),
            'accuracies': np.random.uniform(0.7, 0.95, 30).tolist(),
            'attack_counts': np.random.randint(50, 200, 30).tolist()
        }
    
    def _get_all_models(self) -> List[str]:
        """Get all model IDs"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        cursor.execute("SELECT DISTINCT model_id FROM model_metrics ORDER BY model_id")
        models = [row[0] for row in cursor.fetchall()]
        conn.close()
        return models if models else ['dqn_model_1', 'gan_model_1', 'transformer_model_1']
    
    def _get_model_info(self, model_id: str) -> Dict[str, Any]:
        """Get model information"""
        # Simulate model info for demo
        return {
            'type': 'DQN' if 'dqn' in model_id else 'GAN' if 'gan' in model_id else 'Transformer',
            'predictions': np.random.randint(1000, 10000),
            'success_rate': np.random.uniform(0.6, 0.9),
            'avg_confidence': np.random.uniform(0.7, 0.95),
            'avg_response': np.random.uniform(0.5, 2.0),
            'last_updated': datetime.now().strftime('%Y-%m-%d %H:%M')
        }
    
    def _get_model_metrics(self, model_id: str) -> Dict[str, List[Any]]:
        """Get model metrics history"""
        # Simulate metrics for demo
        timestamps = pd.date_range(end=datetime.now(), periods=100, freq='H')
        
        return {
            'timestamps': timestamps.tolist(),
            'loss': np.random.uniform(0.1, 0.5, 100).tolist(),
            'accuracy': np.random.uniform(0.7, 0.95, 100).tolist(),
            'success_rate': np.random.uniform(0.6, 0.9, 100).tolist(),
            'response_time': np.random.uniform(0.5, 2.0, 100).tolist()
        }


def run_dashboard():
    """Run the ML dashboard"""
    dashboard = MLDashboard()
    dashboard.run()


if __name__ == "__main__":
    run_dashboard()