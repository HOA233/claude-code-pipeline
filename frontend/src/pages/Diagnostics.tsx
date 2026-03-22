import React, { useState, useEffect, useCallback } from 'react';
import api from '../api/client';
import './Diagnostics.css';

interface ComponentHealth {
  name: string;
  status: 'healthy' | 'degraded' | 'unhealthy';
  message: string;
  details?: Record<string, any>;
  last_check: string;
}

interface SystemHealth {
  status: string;
  components: ComponentHealth[];
  uptime: number;
  version: string;
  cpu_usage?: number;
  memory_usage?: number;
  goroutines?: number;
  active_executions?: number;
  total_executions?: number;
}

interface SystemMetric {
  name: string;
  value: number;
  unit: string;
  trend?: 'up' | 'down' | 'stable';
}

function Diagnostics() {
  const [health, setHealth] = useState<SystemHealth | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());

  const fetchHealth = useCallback(async () => {
    setRefreshing(true);
    try {
      const data = await api.getHealthStatus();
      setHealth(data);
      setLastUpdate(new Date());
    } catch (error) {
      console.error('Failed to fetch health:', error);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => {
    fetchHealth();

    if (autoRefresh) {
      const interval = setInterval(fetchHealth, 10000);
      return () => clearInterval(interval);
    }
  }, [fetchHealth, autoRefresh]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
        return '#52c41a';
      case 'degraded':
        return '#faad14';
      case 'unhealthy':
        return '#f5222d';
      default:
        return '#8c8c8c';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'healthy':
        return '✓';
      case 'degraded':
        return '⚠';
      case 'unhealthy':
        return '✗';
      default:
        return '?';
    }
  };

  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    if (days > 0) return `${days}d ${hours}h ${mins}m`;
    if (hours > 0) return `${hours}h ${mins}m`;
    return `${mins}m`;
  };

  const metrics: SystemMetric[] = health ? [
    { name: 'CPU 使用率', value: health.cpu_usage || 0, unit: '%', trend: 'stable' },
    { name: '内存使用率', value: health.memory_usage || 0, unit: '%', trend: 'stable' },
    { name: 'Goroutines', value: health.goroutines || 0, unit: '', trend: 'stable' },
    { name: '活跃执行', value: health.active_executions || 0, unit: '', trend: 'stable' },
  ] : [];

  return (
    <div className="diagnostics-page">
      <div className="page-header">
        <div className="header-left">
          <h1>系统诊断</h1>
          <p>监控系统健康状态和运行指标</p>
        </div>
        <div className="header-right">
          <label className="auto-refresh-toggle">
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
            />
            <span>自动刷新</span>
          </label>
          <button
            className="btn-refresh"
            onClick={fetchHealth}
            disabled={refreshing}
          >
            {refreshing ? '刷新中...' : '刷新'}
          </button>
        </div>
      </div>

      {loading ? (
        <div className="loading">加载中...</div>
      ) : (
        <>
          {/* Overall Status */}
          <div className={`overall-status ${health?.status || 'unknown'}`}>
            <div className="status-icon">
              {getStatusIcon(health?.status || '')}
            </div>
            <div className="status-info">
              <h2>系统状态: {health?.status?.toUpperCase() || 'UNKNOWN'}</h2>
              <p>
                运行时间: {formatUptime(health?.uptime || 0)} |
                版本: {health?.version || 'N/A'} |
                更新: {lastUpdate.toLocaleTimeString('zh-CN')}
              </p>
            </div>
          </div>

          {/* Metrics Grid */}
          <div className="metrics-grid">
            {metrics.map((metric) => (
              <div key={metric.name} className="metric-card">
                <div className="metric-name">{metric.name}</div>
                <div className="metric-value">
                  {metric.value}
                  <span className="metric-unit">{metric.unit}</span>
                </div>
                <div className="metric-bar">
                  {metric.unit === '%' && (
                    <div
                      className="metric-fill"
                      style={{
                        width: `${Math.min(metric.value, 100)}%`,
                        backgroundColor: metric.value > 80 ? '#f5222d' : metric.value > 60 ? '#faad14' : '#52c41a',
                      }}
                    />
                  )}
                </div>
              </div>
            ))}
          </div>

          {/* Components */}
          <div className="components-section">
            <h3>组件状态</h3>
            <div className="components-grid">
              {health?.components?.map((component) => (
                <div
                  key={component.name}
                  className={`component-card ${component.status}`}
                >
                  <div className="component-header">
                    <span className="component-name">{component.name}</span>
                    <span
                      className="component-status"
                      style={{ backgroundColor: getStatusColor(component.status) }}
                    >
                      {component.status}
                    </span>
                  </div>
                  <p className="component-message">{component.message}</p>
                  {component.details && Object.keys(component.details).length > 0 && (
                    <details className="component-details">
                      <summary>详细信息</summary>
                      <pre>{JSON.stringify(component.details, null, 2)}</pre>
                    </details>
                  )}
                  <div className="component-time">
                    最后检查: {new Date(component.last_check).toLocaleString('zh-CN')}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Quick Actions */}
          <div className="quick-actions-section">
            <h3>快速操作</h3>
            <div className="actions-grid">
              <button
                className="action-card"
                onClick={() => api.getSystemMetrics()}
              >
                <span className="action-icon">📊</span>
                <span className="action-label">获取详细指标</span>
              </button>
              <button
                className="action-card"
                onClick={() => api.getExecutionTrends(7)}
              >
                <span className="action-icon">📈</span>
                <span className="action-label">执行趋势 (7天)</span>
              </button>
              <button
                className="action-card"
                onClick={() => window.open('/api/stats/health', '_blank')}
              >
                <span className="action-icon">🔗</span>
                <span className="action-label">原始 API 数据</span>
              </button>
              <button
                className="action-card warning"
                onClick={() => {
                  if (confirm('确定要清除所有执行缓存吗？')) {
                    // Add cache clear logic
                  }
                }}
              >
                <span className="action-icon">🗑️</span>
                <span className="action-label">清除缓存</span>
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  );
}

export default Diagnostics;