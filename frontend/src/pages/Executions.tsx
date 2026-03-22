import React, { useState, useEffect, useCallback } from 'react';
import api from '../api/client';
import type { Execution } from '../types';
import { ExecutionDetailModal } from '../components/ExecutionDetailModal';
import { useToast } from '../components/Toast';
import './Executions.css';

function Executions() {
  const [executions, setExecutions] = useState<Execution[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [selectedExecutionId, setSelectedExecutionId] = useState<string | null>(null);
  const [filter, setFilter] = useState({
    status: '',
    workflow_id: '',
  });
  const [stats, setStats] = useState({
    total: 0,
    running: 0,
    completed: 0,
    failed: 0,
  });
  const { addToast } = useToast();

  const fetchExecutions = useCallback(async () => {
    setLoading(true);
    try {
      const result = await api.listExecutions({
        status: filter.status || undefined,
        workflow_id: filter.workflow_id || undefined,
      });
      const execs = result.executions || [];
      setExecutions(execs);
      setStats({
        total: result.total || execs.length,
        running: execs.filter((e: Execution) => e.status === 'running').length,
        completed: execs.filter((e: Execution) => e.status === 'completed').length,
        failed: execs.filter((e: Execution) => e.status === 'failed').length,
      });
    } catch (error) {
      console.error('Failed to fetch executions:', error);
    } finally {
      setLoading(false);
    }
  }, [filter]);

  useEffect(() => {
    fetchExecutions();

    // Subscribe to real-time updates
    const unsubscribe = api.subscribeAllExecutions(() => {
      fetchExecutions();
    });

    return unsubscribe;
  }, [fetchExecutions]);

  const handleSelectAll = () => {
    if (selectedIds.size === executions.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(executions.map((e) => e.id)));
    }
  };

  const handleSelect = (id: string) => {
    const newSelected = new Set(selectedIds);
    if (newSelected.has(id)) {
      newSelected.delete(id);
    } else {
      newSelected.add(id);
    }
    setSelectedIds(newSelected);
  };

  const handleCancelSelected = async () => {
    if (selectedIds.size === 0) return;
    if (!confirm(`确定要取消 ${selectedIds.size} 个执行吗？`)) return;

    try {
      await Promise.all(
        Array.from(selectedIds).map((id) => api.cancelExecution(id))
      );
      addToast(`已取消 ${selectedIds.size} 个执行`, 'success');
      setSelectedIds(new Set());
      fetchExecutions();
    } catch (error) {
      console.error('Failed to cancel executions:', error);
      addToast('取消失败', 'error');
    }
  };

  const handleCancelAllRunning = async () => {
    if (!confirm('确定要取消所有运行中的执行吗？')) return;
    try {
      const result = await api.cancelAllExecutions('running');
      addToast(`已取消 ${result.count} 个执行`, 'success');
      fetchExecutions();
    } catch (error) {
      console.error('Failed to cancel all:', error);
      addToast('取消失败', 'error');
    }
  };

  const handleRetry = async (id: string) => {
    try {
      await api.retryExecution(id);
      addToast('已重新执行', 'success');
      fetchExecutions();
    } catch (error) {
      console.error('Failed to retry:', error);
      addToast('重试失败', 'error');
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running': return '#1890ff';
      case 'completed': return '#52c41a';
      case 'failed': return '#f5222d';
      case 'paused': return '#faad14';
      case 'cancelled': return '#8c8c8c';
      default: return '#8c8c8c';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'running': return '▶';
      case 'completed': return '✓';
      case 'failed': return '✗';
      case 'paused': return '⏸';
      case 'cancelled': return '⏹';
      default: return '○';
    }
  };

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${Math.round(ms / 1000)}s`;
    return `${Math.round(ms / 60000)}m`;
  };

  return (
    <div className="executions-page">
      <div className="page-header">
        <div className="header-left">
          <h1>任务执行</h1>
          <p>查看和管理所有任务执行状态</p>
        </div>
        <div className="header-right">
          <button className="btn-refresh" onClick={fetchExecutions}>
            刷新
          </button>
          <button className="btn-danger" onClick={handleCancelAllRunning}>
            取消所有运行中
          </button>
        </div>
      </div>

      {/* Stats */}
      <div className="stats-bar">
        <div className="stat-item">
          <span className="stat-label">总计</span>
          <span className="stat-value">{stats.total}</span>
        </div>
        <div className="stat-item running">
          <span className="stat-label">运行中</span>
          <span className="stat-value">{stats.running}</span>
        </div>
        <div className="stat-item completed">
          <span className="stat-label">已完成</span>
          <span className="stat-value">{stats.completed}</span>
        </div>
        <div className="stat-item failed">
          <span className="stat-label">失败</span>
          <span className="stat-value">{stats.failed}</span>
        </div>
      </div>

      {/* Filters */}
      <div className="filters-bar">
        <select
          value={filter.status}
          onChange={(e) => setFilter({ ...filter, status: e.target.value })}
        >
          <option value="">全部状态</option>
          <option value="running">运行中</option>
          <option value="completed">已完成</option>
          <option value="failed">失败</option>
          <option value="paused">已暂停</option>
          <option value="cancelled">已取消</option>
        </select>

        {selectedIds.size > 0 && (
          <div className="bulk-actions">
            <span className="selected-count">已选择 {selectedIds.size} 项</span>
            <button className="btn-danger" onClick={handleCancelSelected}>
              取消选中
            </button>
            <button onClick={() => setSelectedIds(new Set())}>
              取消选择
            </button>
          </div>
        )}
      </div>

      {/* Table */}
      {loading ? (
        <div className="loading">加载中...</div>
      ) : (
        <div className="executions-table-container">
          <table className="executions-table">
            <thead>
              <tr>
                <th className="th-checkbox">
                  <input
                    type="checkbox"
                    checked={selectedIds.size === executions.length && executions.length > 0}
                    onChange={handleSelectAll}
                  />
                </th>
                <th>状态</th>
                <th>工作流</th>
                <th>进度</th>
                <th>步骤</th>
                <th>耗时</th>
                <th>创建时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {executions.map((execution) => (
                <tr
                  key={execution.id}
                  className={selectedIds.has(execution.id) ? 'selected' : ''}
                >
                  <td>
                    <input
                      type="checkbox"
                      checked={selectedIds.has(execution.id)}
                      onChange={() => handleSelect(execution.id)}
                    />
                  </td>
                  <td>
                    <span
                      className="status-badge"
                      style={{ backgroundColor: getStatusColor(execution.status) }}
                    >
                      {getStatusIcon(execution.status)} {execution.status}
                    </span>
                  </td>
                  <td
                    className="workflow-name"
                    onClick={() => setSelectedExecutionId(execution.id)}
                  >
                    {execution.workflow_name}
                  </td>
                  <td>
                    <div className="progress-cell">
                      <div className="progress-bar">
                        <div
                          className="progress-fill"
                          style={{ width: `${execution.progress}%` }}
                        />
                      </div>
                      <span className="progress-text">{execution.progress}%</span>
                    </div>
                  </td>
                  <td>{execution.completed_steps}/{execution.total_steps}</td>
                  <td>{formatDuration(execution.duration)}</td>
                  <td className="time-cell">
                    {new Date(execution.created_at).toLocaleString('zh-CN')}
                  </td>
                  <td>
                    <div className="row-actions">
                      <button onClick={() => setSelectedExecutionId(execution.id)}>
                        详情
                      </button>
                      {execution.status === 'running' && (
                        <button
                          className="btn-warning"
                          onClick={async () => {
                            await api.cancelExecution(execution.id);
                            fetchExecutions();
                          }}
                        >
                          取消
                        </button>
                      )}
                      {execution.status === 'failed' && (
                        <button onClick={() => handleRetry(execution.id)}>
                          重试
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {executions.length === 0 && (
            <div className="empty-state">
              <span className="empty-icon">📋</span>
              <p>暂无执行记录</p>
            </div>
          )}
        </div>
      )}

      {selectedExecutionId && (
        <ExecutionDetailModal
          executionId={selectedExecutionId}
          onClose={() => setSelectedExecutionId(null)}
        />
      )}
    </div>
  );
}

export default Executions;