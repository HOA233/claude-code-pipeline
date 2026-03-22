import React, { useState, useEffect, useCallback } from 'react';
import api from '../api/client';
import './SystemStatusIndicator.css';

interface StatusData {
  status: 'online' | 'degraded' | 'offline';
  version: string;
  uptime: number;
  lastCheck: Date;
  message?: string;
}

interface SystemStatusIndicatorProps {
  showDetails?: boolean;
  refreshInterval?: number;
}

export const SystemStatusIndicator: React.FC<SystemStatusIndicatorProps> = ({
  showDetails = false,
  refreshInterval = 30000,
}) => {
  const [status, setStatus] = useState<StatusData>({
    status: 'online',
    version: '1.0.0',
    uptime: 0,
    lastCheck: new Date(),
  });
  const [expanded, setExpanded] = useState(false);
  const [loading, setLoading] = useState(true);

  const fetchStatus = useCallback(async () => {
    try {
      const result = await api.getHealthStatus();
      setStatus({
        status: result.status || 'online',
        version: result.version || '1.0.0',
        uptime: result.uptime || 0,
        lastCheck: new Date(),
        message: result.message,
      });
    } catch (error) {
      console.error('Failed to fetch system status:', error);
      setStatus((prev) => ({
        ...prev,
        status: 'offline',
        lastCheck: new Date(),
        message: '无法连接到服务器',
      }));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchStatus();
    const interval = setInterval(fetchStatus, refreshInterval);
    return () => clearInterval(interval);
  }, [fetchStatus, refreshInterval]);

  const getStatusColor = () => {
    switch (status.status) {
      case 'online':
        return '#52c41a';
      case 'degraded':
        return '#faad14';
      case 'offline':
        return '#f5222d';
      default:
        return '#8c8c8c';
    }
  };

  const getStatusText = () => {
    switch (status.status) {
      case 'online':
        return '系统正常';
      case 'degraded':
        return '性能下降';
      case 'offline':
        return '系统离线';
      default:
        return '未知状态';
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

  return (
    <div className={`system-status-indicator ${expanded ? 'expanded' : ''}`}>
      <div
        className="status-summary"
        onClick={() => showDetails && setExpanded(!expanded)}
      >
        <div
          className="status-dot"
          style={{ backgroundColor: getStatusColor() }}
        />
        <span className="status-text">{getStatusText()}</span>
        {showDetails && (
          <span className="expand-icon">{expanded ? '▲' : '▼'}</span>
        )}
      </div>

      {expanded && showDetails && (
        <div className="status-details">
          <div className="detail-row">
            <span className="detail-label">状态</span>
            <span className="detail-value" style={{ color: getStatusColor() }}>
              {status.status.toUpperCase()}
            </span>
          </div>
          <div className="detail-row">
            <span className="detail-label">版本</span>
            <span className="detail-value">{status.version}</span>
          </div>
          <div className="detail-row">
            <span className="detail-label">运行时间</span>
            <span className="detail-value">{formatUptime(status.uptime)}</span>
          </div>
          <div className="detail-row">
            <span className="detail-label">最后检查</span>
            <span className="detail-value">
              {status.lastCheck.toLocaleTimeString('zh-CN')}
            </span>
          </div>
          {status.message && (
            <div className="detail-message">{status.message}</div>
          )}
          <button className="refresh-btn" onClick={fetchStatus} disabled={loading}>
            {loading ? '检查中...' : '刷新状态'}
          </button>
        </div>
      )}
    </div>
  );
};

export default SystemStatusIndicator;