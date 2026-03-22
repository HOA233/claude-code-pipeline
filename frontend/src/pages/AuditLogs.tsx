import React, { useState, useEffect, useCallback } from 'react';
import api from '../api/client';
import './AuditLogs.css';

interface AuditLog {
  id: string;
  action: string;
  resource: string;
  resource_id: string;
  actor: string;
  actor_type: string;
  ip_address?: string;
  user_agent?: string;
  details?: Record<string, any>;
  status: string;
  timestamp: string;
}

function AuditLogs() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [filter, setFilter] = useState({
    action: '',
    resource: '',
    actor: '',
  });

  const fetchLogs = useCallback(async () => {
    setLoading(true);
    try {
      const params: any = {};
      if (filter.action) params.action = filter.action;
      if (filter.resource) params.resource = filter.resource;
      if (filter.actor) params.actor = filter.actor;

      const result = await api.listAuditLogs(params);
      setLogs(result.logs || []);
      setTotal(result.total || 0);
    } catch (error) {
      console.error('Failed to fetch audit logs:', error);
    } finally {
      setLoading(false);
    }
  }, [filter]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  const getActionColor = (action: string) => {
    switch (action) {
      case 'create': return '#52c41a';
      case 'update': return '#1890ff';
      case 'delete': return '#f5222d';
      case 'execute': return '#722ed1';
      default: return '#8c8c8c';
    }
  };

  const getResourceIcon = (type: string) => {
    switch (type) {
      case 'agent': return '🤖';
      case 'workflow': return '🔄';
      case 'schedule': return '⏰';
      case 'webhook': return '🔔';
      case 'execution': return '📊';
      default: return '📄';
    }
  };

  const formatDate = (timestamp: string) => {
    return new Date(timestamp).toLocaleString('zh-CN');
  };

  const handleExport = async (format: 'json' | 'csv') => {
    await api.exportAuditLogs(format);
  };

  return (
    <div className="audit-logs-page">
      <div className="page-header">
        <h1>审计日志</h1>
        <p>查看系统操作记录和变更历史</p>
      </div>

      <div className="filters-bar">
        <select
          value={filter.action}
          onChange={(e) => setFilter({ ...filter, action: e.target.value })}
        >
          <option value="">全部操作</option>
          <option value="create">创建</option>
          <option value="update">更新</option>
          <option value="delete">删除</option>
          <option value="execute">执行</option>
        </select>

        <select
          value={filter.resource}
          onChange={(e) => setFilter({ ...filter, resource: e.target.value })}
        >
          <option value="">全部资源</option>
          <option value="agent">Agent</option>
          <option value="workflow">Workflow</option>
          <option value="schedule">Schedule</option>
          <option value="webhook">Webhook</option>
          <option value="execution">Execution</option>
        </select>

        <button className="btn-refresh" onClick={fetchLogs}>
          刷新
        </button>

        <button className="btn-export" onClick={() => handleExport('json')}>
          导出 JSON
        </button>

        <button className="btn-export" onClick={() => handleExport('csv')}>
          导出 CSV
        </button>
      </div>

      {loading ? (
        <div className="loading">加载中...</div>
      ) : (
        <div className="logs-table-container">
          <table className="logs-table">
            <thead>
              <tr>
                <th>时间</th>
                <th>操作</th>
                <th>资源</th>
                <th>资源 ID</th>
                <th>操作者</th>
                <th>IP 地址</th>
                <th>状态</th>
                <th>详情</th>
              </tr>
            </thead>
            <tbody>
              {logs.map((log) => (
                <tr key={log.id} className={log.status === 'failed' ? 'row-failure' : ''}>
                  <td className="td-time">{formatDate(log.timestamp)}</td>
                  <td>
                    <span
                      className="action-badge"
                      style={{ backgroundColor: getActionColor(log.action) }}
                    >
                      {log.action}
                    </span>
                  </td>
                  <td>
                    <span className="resource-type">
                      {getResourceIcon(log.resource)} {log.resource}
                    </span>
                  </td>
                  <td className="td-id">{log.resource_id}</td>
                  <td>
                    <span className={`actor-badge ${log.actor_type}`}>
                      {log.actor_type === 'user' ? '👤' : log.actor_type === 'api' ? '🔑' : '⚙️'} {log.actor}
                    </span>
                  </td>
                  <td className="td-ip">{log.ip_address || '-'}</td>
                  <td>
                    <span className={`status-badge ${log.status}`}>
                      {log.status === 'success' ? '✓ 成功' : '✗ 失败'}
                    </span>
                  </td>
                  <td>
                    {log.details && Object.keys(log.details).length > 0 && (
                      <details className="details-dropdown">
                        <summary>查看</summary>
                        <pre>{JSON.stringify(log.details, null, 2)}</pre>
                      </details>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {logs.length === 0 && (
            <div className="empty-logs">
              <span className="empty-icon">📋</span>
              <p>暂无审计日志</p>
            </div>
          )}
        </div>
      )}

      <div className="pagination">
        <span className="total-count">共 {total} 条记录</span>
      </div>
    </div>
  );
}

export default AuditLogs;