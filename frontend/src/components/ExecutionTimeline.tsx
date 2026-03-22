import React, { useState, useEffect, useCallback } from 'react';
import api from '../api/client';
import type { Execution } from '../types';
import './ExecutionTimeline.css';

interface ExecutionTimelineProps {
  workflowId?: string;
  limit?: number;
  onSelect?: (execution: Execution) => void;
}

export const ExecutionTimeline: React.FC<ExecutionTimelineProps> = ({
  workflowId,
  limit = 10,
  onSelect,
}) => {
  const [executions, setExecutions] = useState<Execution[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchExecutions = useCallback(async () => {
    setLoading(true);
    try {
      const result = await api.listExecutions({
        workflow_id: workflowId,
        page: 1,
        page_size: limit,
      });
      // Sort by created_at descending
      const sorted = (result.executions || []).sort(
        (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      );
      setExecutions(sorted);
    } catch (error) {
      console.error('Failed to fetch executions:', error);
    } finally {
      setLoading(false);
    }
  }, [workflowId, limit]);

  useEffect(() => {
    fetchExecutions();
  }, [fetchExecutions]);

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return '✓';
      case 'failed':
        return '✗';
      case 'running':
        return '▶';
      case 'paused':
        return '⏸';
      case 'cancelled':
        return '⏹';
      default:
        return '○';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return '#52c41a';
      case 'failed':
        return '#f5222d';
      case 'running':
        return '#1890ff';
      case 'paused':
        return '#faad14';
      default:
        return '#8c8c8c';
    }
  };

  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return '刚刚';
    if (diffMins < 60) return `${diffMins}分钟前`;
    if (diffHours < 24) return `${diffHours}小时前`;
    if (diffDays < 7) return `${diffDays}天前`;
    return date.toLocaleDateString('zh-CN');
  };

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${Math.round(ms / 1000)}s`;
    return `${Math.round(ms / 60000)}m`;
  };

  if (loading) {
    return <div className="execution-timeline loading">加载中...</div>;
  }

  if (executions.length === 0) {
    return <div className="execution-timeline empty">暂无执行记录</div>;
  }

  return (
    <div className="execution-timeline">
      <div className="timeline-header">
        <h4>执行历史</h4>
        <span className="timeline-count">{executions.length} 条记录</span>
      </div>

      <div className="timeline-list">
        {executions.map((exec, index) => (
          <div
            key={exec.id}
            className={`timeline-item timeline-item-${exec.status}`}
            onClick={() => onSelect?.(exec)}
          >
            <div className="timeline-marker">
              <div
                className="marker-dot"
                style={{ backgroundColor: getStatusColor(exec.status) }}
              >
                {getStatusIcon(exec.status)}
              </div>
              {index < executions.length - 1 && (
                <div className="marker-line" />
              )}
            </div>

            <div className="timeline-content">
              <div className="timeline-title">
                <span className="workflow-name">{exec.workflow_name}</span>
                <span
                  className="status-badge"
                  style={{ color: getStatusColor(exec.status) }}
                >
                  {exec.status}
                </span>
              </div>

              <div className="timeline-meta">
                <span className="time">{formatTime(exec.created_at)}</span>
                <span className="duration">{formatDuration(exec.duration)}</span>
                <span className="progress">{exec.progress}%</span>
              </div>

              <div className="timeline-progress">
                <div className="progress-bar">
                  <div
                    className="progress-fill"
                    style={{
                      width: `${exec.progress}%`,
                      backgroundColor: getStatusColor(exec.status),
                    }}
                  />
                </div>
              </div>

              {exec.error && (
                <div className="timeline-error">{exec.error}</div>
              )}
            </div>
          </div>
        ))}
      </div>

      <div className="timeline-footer">
        <button className="view-all-btn" onClick={() => window.location.href = '/executions'}>
          查看全部
        </button>
      </div>
    </div>
  );
};

export default ExecutionTimeline;