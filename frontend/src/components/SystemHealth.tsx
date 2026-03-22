import React, { useState, useEffect, useCallback } from 'react';
import api from '../api/client';
import './SystemHealth.css';

interface ComponentHealth {
  status: string;
  latency?: number;
  running_tasks?: number;
  active_jobs?: number;
}

interface HealthStatus {
  status: 'healthy' | 'degraded' | 'unhealthy';
  components: {
    redis: ComponentHealth;
    executor: ComponentHealth;
    scheduler: ComponentHealth;
  };
  uptime?: number;
  version?: string;
  last_check: string;
}

interface SystemMetrics {
  cpu_usage: number;
  memory_usage: number;
  goroutines: number;
  active_executions: number;
  queued_tasks: number;
  redis_memory: number;
  redis_keys: number;
}

export const SystemHealth: React.FC = () => {
  const [health, setHealth] = useState<HealthStatus | null>(null);
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [autoRefresh, setAutoRefresh] = useState(true);

  const fetchData = useCallback(async () => {
    try {
      const [healthRes, metricsRes] = await Promise.all([
        api.getHealthStatus(),
        api.getSystemMetrics(),
      ]);
      setHealth(healthRes);
      setMetrics(metricsRes);
    } catch (error) {
      console.error('Failed to fetch system health:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  useEffect(() => {
    if (!autoRefresh) return;
    const interval = setInterval(fetchData, 10000);
    return () => clearInterval(interval);
  }, [autoRefresh, fetchData]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
      case 'up':
        return '#52c41a';
      case 'degraded':
      case 'warning':
        return '#faad14';
      case 'unhealthy':
      case 'down':
        return '#f5222d';
      default:
        return '#8c8c8c';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'healthy':
      case 'up':
        return '●';
      case 'degraded':
      case 'warning':
        return '◐';
      case 'unhealthy':
      case 'down':
        return '○';
      default:
        return '○';
    }
  };

  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    if (days > 0) return `${days}天 ${hours}小时`;
    if (hours > 0) return `${hours}小时 ${mins}分钟`;
    return `${mins}分钟`;
  };

  if (loading) {
    return <div className="system-health loading">加载中...</div>;
  }

  return (
    <div className="system-health">
      <div className="health-header">
        <h3>系统健康状态</h3>
        <label className="auto-refresh-toggle">
          <input
            type="checkbox"
            checked={autoRefresh}
            onChange={(e) => setAutoRefresh(e.target.checked)}
          />
          自动刷新
        </label>
      </div>

      {/* Overall Status */}
      <div className={`overall-status status-${health?.status || 'unknown'}`}>
        <span className="status-icon">{getStatusIcon(health?.status || '')}</span>
        <span className="status-label">
          {health?.status === 'healthy' ? '系统正常' :
           health?.status === 'degraded' ? '系统降级' :
           health?.status === 'unhealthy' ? '系统异常' : '状态未知'}
        </span>
        {health?.uptime && (
          <span className="uptime">运行时间: {formatUptime(health.uptime)}</span>
        )}
      </div>

      {/* Component Status */}
      <div className="components-grid">
        <div className="component-card">
          <div className="component-header">
            <span className="component-icon">💾</span>
            <span className="component-name">Redis</span>
            <span
              className="component-status"
              style={{ color: getStatusColor(health?.components?.redis?.status || '') }}
            >
              {getStatusIcon(health?.components?.redis?.status || '')}
              {health?.components?.redis?.status || 'unknown'}
            </span>
          </div>
          {health?.components?.redis?.latency !== undefined && (
            <div className="component-metric">
              延迟: {health.components.redis.latency}ms
            </div>
          )}
        </div>

        <div className="component-card">
          <div className="component-header">
            <span className="component-icon">⚡</span>
            <span className="component-name">执行器</span>
            <span
              className="component-status"
              style={{ color: getStatusColor(health?.components?.executor?.status || '') }}
            >
              {getStatusIcon(health?.components?.executor?.status || '')}
              {health?.components?.executor?.status || 'unknown'}
            </span>
          </div>
          {health?.components?.executor?.running_tasks !== undefined && (
            <div className="component-metric">
              运行中任务: {health.components.executor.running_tasks}
            </div>
          )}
        </div>

        <div className="component-card">
          <div className="component-header">
            <span className="component-icon">⏰</span>
            <span className="component-name">调度器</span>
            <span
              className="component-status"
              style={{ color: getStatusColor(health?.components?.scheduler?.status || '') }}
            >
              {getStatusIcon(health?.components?.scheduler?.status || '')}
              {health?.components?.scheduler?.status || 'unknown'}
            </span>
          </div>
          {health?.components?.scheduler?.active_jobs !== undefined && (
            <div className="component-metric">
              活跃任务: {health.components.scheduler.active_jobs}
            </div>
          )}
        </div>
      </div>

      {/* System Metrics */}
      {metrics && (
        <div className="metrics-section">
          <h4>系统指标</h4>
          <div className="metrics-grid">
            <div className="metric-item">
              <div className="metric-label">CPU 使用率</div>
              <div className="metric-bar">
                <div
                  className="metric-fill"
                  style={{
                    width: `${metrics.cpu_usage}%`,
                    backgroundColor: metrics.cpu_usage > 80 ? '#f5222d' :
                                    metrics.cpu_usage > 60 ? '#faad14' : '#52c41a'
                  }}
                />
              </div>
              <div className="metric-value">{metrics.cpu_usage.toFixed(1)}%</div>
            </div>

            <div className="metric-item">
              <div className="metric-label">内存使用率</div>
              <div className="metric-bar">
                <div
                  className="metric-fill"
                  style={{
                    width: `${metrics.memory_usage}%`,
                    backgroundColor: metrics.memory_usage > 80 ? '#f5222d' :
                                    metrics.memory_usage > 60 ? '#faad14' : '#52c41a'
                  }}
                />
              </div>
              <div className="metric-value">{metrics.memory_usage.toFixed(1)}%</div>
            </div>

            <div className="metric-item">
              <div className="metric-label">Goroutines</div>
              <div className="metric-value">{metrics.goroutines}</div>
            </div>

            <div className="metric-item">
              <div className="metric-label">活跃执行</div>
              <div className="metric-value">{metrics.active_executions}</div>
            </div>

            <div className="metric-item">
              <div className="metric-label">队列任务</div>
              <div className="metric-value">{metrics.queued_tasks}</div>
            </div>

            <div className="metric-item">
              <div className="metric-label">Redis 内存</div>
              <div className="metric-value">
                {metrics.redis_memory > 1024 * 1024
                  ? `${(metrics.redis_memory / 1024 / 1024).toFixed(1)}MB`
                  : `${(metrics.redis_memory / 1024).toFixed(1)}KB`}
              </div>
            </div>
          </div>
        </div>
      )}

      <div className="health-footer">
        <span className="last-check">
          最后检查: {health?.last_check ? new Date(health.last_check).toLocaleString('zh-CN') : '-'}
        </span>
        <button onClick={fetchData} className="refresh-btn">立即刷新</button>
      </div>
    </div>
  );
};

export default SystemHealth;