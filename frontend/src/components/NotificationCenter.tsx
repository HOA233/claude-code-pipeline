import React, { useState, useEffect, useCallback } from 'react';
import './NotificationCenter.css';

interface Notification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message: string;
  timestamp: string;
  read: boolean;
  source?: string;
  link?: string;
}

interface NotificationCenterProps {
  maxItems?: number;
}

export const NotificationCenter: React.FC<NotificationCenterProps> = ({
  maxItems = 50,
}) => {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [filter, setFilter] = useState<'all' | 'unread'>('all');
  const [typeFilter, setTypeFilter] = useState<string>('all');

  const unreadCount = notifications.filter((n) => !n.read).length;

  const addNotification = useCallback((notification: Omit<Notification, 'id' | 'timestamp' | 'read'>) => {
    const newNotification: Notification = {
      ...notification,
      id: Date.now().toString(),
      timestamp: new Date().toISOString(),
      read: false,
    };
    setNotifications((prev) => [newNotification, ...prev].slice(0, maxItems));
  }, [maxItems]);

  // Simulate some initial notifications
  useEffect(() => {
    setNotifications([
      {
        id: '1',
        type: 'success',
        title: '工作流执行成功',
        message: 'code-review-workflow 已完成执行',
        timestamp: new Date(Date.now() - 3600000).toISOString(),
        read: false,
        source: 'workflow',
        link: '/executions',
      },
      {
        id: '2',
        type: 'warning',
        title: '系统资源警告',
        message: 'CPU 使用率超过 80%',
        timestamp: new Date(Date.now() - 7200000).toISOString(),
        read: false,
        source: 'system',
        link: '/diagnostics',
      },
      {
        id: '3',
        type: 'info',
        title: '新 Agent 已创建',
        message: 'test-generator 已成功创建',
        timestamp: new Date(Date.now() - 86400000).toISOString(),
        read: true,
        source: 'agent',
        link: '/agents',
      },
    ]);
  }, []);

  const markAsRead = (id: string) => {
    setNotifications((prev) =>
      prev.map((n) => (n.id === id ? { ...n, read: true } : n))
    );
  };

  const markAllAsRead = () => {
    setNotifications((prev) => prev.map((n) => ({ ...n, read: true })));
  };

  const clearAll = () => {
    if (confirm('确定要清除所有通知吗？')) {
      setNotifications([]);
    }
  };

  const deleteNotification = (id: string) => {
    setNotifications((prev) => prev.filter((n) => n.id !== id));
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'success': return '✓';
      case 'warning': return '⚠';
      case 'error': return '✗';
      default: return 'ℹ';
    }
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'success': return '#52c41a';
      case 'warning': return '#faad14';
      case 'error': return '#f5222d';
      default: return '#1890ff';
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

  const filteredNotifications = notifications.filter((n) => {
    if (filter === 'unread' && n.read) return false;
    if (typeFilter !== 'all' && n.type !== typeFilter) return false;
    return true;
  });

  return (
    <div className="notification-center">
      <button
        className="notification-bell"
        onClick={() => setIsOpen(!isOpen)}
      >
        🔔
        {unreadCount > 0 && (
          <span className="badge">{unreadCount > 99 ? '99+' : unreadCount}</span>
        )}
      </button>

      {isOpen && (
        <div className="notification-dropdown">
          <div className="dropdown-header">
            <h3>通知中心</h3>
            <div className="header-actions">
              {unreadCount > 0 && (
                <button onClick={markAllAsRead}>全部已读</button>
              )}
              <button onClick={clearAll}>清空</button>
            </div>
          </div>

          <div className="dropdown-filters">
            <select value={filter} onChange={(e) => setFilter(e.target.value as any)}>
              <option value="all">全部</option>
              <option value="unread">未读</option>
            </select>
            <select value={typeFilter} onChange={(e) => setTypeFilter(e.target.value)}>
              <option value="all">全部类型</option>
              <option value="success">成功</option>
              <option value="warning">警告</option>
              <option value="error">错误</option>
              <option value="info">信息</option>
            </select>
          </div>

          <div className="dropdown-list">
            {filteredNotifications.length > 0 ? (
              filteredNotifications.map((notification) => (
                <div
                  key={notification.id}
                  className={`notification-item ${notification.read ? 'read' : 'unread'}`}
                  onClick={() => markAsRead(notification.id)}
                >
                  <div
                    className="notification-icon"
                    style={{ color: getTypeColor(notification.type) }}
                  >
                    {getTypeIcon(notification.type)}
                  </div>
                  <div className="notification-content">
                    <div className="notification-title">{notification.title}</div>
                    <div className="notification-message">{notification.message}</div>
                    <div className="notification-meta">
                      <span className="notification-time">
                        {formatTime(notification.timestamp)}
                      </span>
                      {notification.source && (
                        <span className="notification-source">
                          {notification.source}
                        </span>
                      )}
                    </div>
                  </div>
                  <button
                    className="notification-delete"
                    onClick={(e) => {
                      e.stopPropagation();
                      deleteNotification(notification.id);
                    }}
                  >
                    ×
                  </button>
                </div>
              ))
            ) : (
              <div className="empty-notifications">
                <span className="empty-icon">📭</span>
                <p>暂无通知</p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default NotificationCenter;