import React, { useState, useEffect, useCallback } from 'react';
import api from '../api/client';
import type { Execution } from '../types';
import './ActivityFeed.css';

interface Activity {
  id: string;
  type: 'execution_started' | 'execution_completed' | 'execution_failed' | 'agent_created' | 'workflow_created' | 'schedule_triggered' | 'webhook_delivered';
  title: string;
  description: string;
  timestamp: string;
  metadata?: Record<string, any>;
}

interface ActivityFeedProps {
  limit?: number;
  showHeader?: boolean;
  maxHeight?: string;
}

export const ActivityFeed: React.FC<ActivityFeedProps> = ({
  limit = 20,
  showHeader = true,
  maxHeight = '400px',
}) => {
  const [activities, setActivities] = useState<Activity[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<string>('all');

  const fetchActivities = useCallback(async () => {
    setLoading(true);
    try {
      // Get recent executions to build activity feed
      const execResult = await api.listExecutions({ page: 1, page_size: limit });
      const executions = execResult.executions || [];

      // Convert executions to activities
      const execActivities: Activity[] = executions.map((exec: Execution) => {
        let type: Activity['type'] = 'execution_started';
        let title = '';
        let description = '';

        switch (exec.status) {
          case 'running':
            type = 'execution_started';
            title = `开始执行: ${exec.workflow_name}`;
            description = `工作流 ${exec.workflow_name} 开始执行`;
            break;
          case 'completed':
            type = 'execution_completed';
            title = `执行完成: ${exec.workflow_name}`;
            description = `工作流 ${exec.workflow_name} 已成功完成`;
            break;
          case 'failed':
            type = 'execution_failed';
            title = `执行失败: ${exec.workflow_name}`;
            description = exec.error || `工作流 ${exec.workflow_name} 执行失败`;
            break;
          case 'cancelled':
            type = 'execution_failed';
            title = `执行取消: ${exec.workflow_name}`;
            description = `工作流 ${exec.workflow_name} 已被取消`;
            break;
          default:
            type = 'execution_started';
            title = `执行中: ${exec.workflow_name}`;
            description = `工作流 ${exec.workflow_name} 正在执行`;
        }

        return {
          id: exec.id,
          type,
          title,
          description,
          timestamp: exec.created_at,
          metadata: {
            workflow_id: exec.workflow_id,
            workflow_name: exec.workflow_name,
            duration: exec.duration,
            progress: exec.progress,
            status: exec.status,
          },
        };
      });

      setActivities(execActivities);
    } catch (error) {
      console.error('Failed to fetch activities:', error);
    } finally {
      setLoading(false);
    }
  }, [limit]);

  useEffect(() => {
    fetchActivities();

    // Subscribe to real-time updates
    const unsubscribe = api.subscribeAllExecutions(() => {
      fetchActivities();
    });

    return unsubscribe;
  }, [fetchActivities]);

  const getActivityIcon = (type: Activity['type']) => {
    switch (type) {
      case 'execution_started':
        return '▶️';
      case 'execution_completed':
        return '✅';
      case 'execution_failed':
        return '❌';
      case 'agent_created':
        return '🤖';
      case 'workflow_created':
        return '🔄';
      case 'schedule_triggered':
        return '⏰';
      case 'webhook_delivered':
        return '🔔';
      default:
        return '📝';
    }
  };

  const getActivityColor = (type: Activity['type']) => {
    switch (type) {
      case 'execution_started':
        return '#1890ff';
      case 'execution_completed':
        return '#52c41a';
      case 'execution_failed':
        return '#f5222d';
      case 'agent_created':
        return '#722ed1';
      case 'workflow_created':
        return '#13c2c2';
      case 'schedule_triggered':
        return '#faad14';
      case 'webhook_delivered':
        return '#eb2f96';
      default:
        return '#8c8c8c';
    }
  };

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now.getTime() - date.getTime();

    if (diff < 60000) return '刚刚';
    if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`;
    return date.toLocaleDateString('zh-CN');
  };

  const filteredActivities = activities.filter((activity) => {
    if (filter === 'all') return true;
    return activity.type.includes(filter);
  });

  return (
    <div className="activity-feed" style={{ maxHeight }}>
      {showHeader && (
        <div className="feed-header">
          <h3>实时动态</h3>
          <div className="feed-actions">
            <select
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              className="filter-select"
            >
              <option value="all">全部</option>
              <option value="execution">执行</option>
              <option value="completed">完成</option>
              <option value="failed">失败</option>
            </select>
            <button className="refresh-btn" onClick={fetchActivities} disabled={loading}>
              {loading ? '...' : '刷新'}
            </button>
          </div>
        </div>
      )}

      <div className="feed-content">
        {loading && activities.length === 0 ? (
          <div className="feed-loading">加载中...</div>
        ) : filteredActivities.length === 0 ? (
          <div className="feed-empty">
            <span className="empty-icon">📭</span>
            <span>暂无动态</span>
          </div>
        ) : (
          <div className="activity-list">
            {filteredActivities.map((activity) => (
              <div key={activity.id} className="activity-item">
                <div
                  className="activity-icon"
                  style={{ backgroundColor: `${getActivityColor(activity.type)}20` }}
                >
                  <span>{getActivityIcon(activity.type)}</span>
                </div>
                <div className="activity-content">
                  <div className="activity-title">{activity.title}</div>
                  <div className="activity-description">{activity.description}</div>
                  {activity.metadata && (
                    <div className="activity-meta">
                      {activity.metadata.duration && (
                        <span className="meta-item">
                          耗时: {activity.metadata.duration}ms
                        </span>
                      )}
                      {activity.metadata.progress !== undefined && activity.metadata.progress < 100 && (
                        <span className="meta-item">
                          进度: {activity.metadata.progress}%
                        </span>
                      )}
                    </div>
                  )}
                </div>
                <div className="activity-time">{formatTime(activity.timestamp)}</div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default ActivityFeed;