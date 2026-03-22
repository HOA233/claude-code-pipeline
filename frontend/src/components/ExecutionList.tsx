import React, { useState, useEffect, useCallback } from 'react';
import type { Execution, ExecutionStatus } from '../types';
import api from '../api/client';
import { ExecutionDetailModal } from './ExecutionDetailModal';
import { useToast } from './Toast';

interface ExecutionListProps {
  tenantId?: string;
  autoRefresh?: boolean;
  refreshInterval?: number;
}

export const ExecutionList: React.FC<ExecutionListProps> = ({
  tenantId,
  autoRefresh = true,
  refreshInterval = 5000,
}) => {
  const [executions, setExecutions] = useState<Execution[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [selectedStatus, setSelectedStatus] = useState<ExecutionStatus | ''>('');
  const [selectedExecutionId, setSelectedExecutionId] = useState<string | null>(null);
  const [retryingId, setRetryingId] = useState<string | null>(null);
  const { addToast } = useToast();

  const fetchExecutions = useCallback(async () => {
    setLoading(true);
    try {
      const result = await api.listExecutions({
        tenant_id: tenantId,
        status: selectedStatus || undefined,
        page,
        page_size: 20,
      });
      setExecutions(result.executions);
      setTotal(result.total);
    } catch (error) {
      console.error('Failed to fetch executions:', error);
    } finally {
      setLoading(false);
    }
  }, [tenantId, selectedStatus, page]);

  useEffect(() => {
    fetchExecutions();
  }, [fetchExecutions]);

  useEffect(() => {
    if (!autoRefresh) return;

    const unsubscribe = api.subscribeAllExecutions((data) => {
      setExecutions((prev) =>
        prev.map((exec) =>
          exec.id === data.execution_id
            ? { ...exec, status: data.status, progress: data.progress }
            : exec
        )
      );
    });

    return unsubscribe;
  }, [autoRefresh]);

  const handleCancel = async (id: string) => {
    try {
      await api.cancelExecution(id);
      fetchExecutions();
    } catch (error) {
      console.error('Failed to cancel execution:', error);
    }
  };

  const handlePause = async (id: string) => {
    try {
      await api.pauseExecution(id);
      fetchExecutions();
    } catch (error) {
      console.error('Failed to pause execution:', error);
    }
  };

  const handleResume = async (id: string) => {
    try {
      await api.resumeExecution(id);
      fetchExecutions();
    } catch (error) {
      console.error('Failed to resume execution:', error);
    }
  };

  const handleCancelAll = async () => {
    try {
      const result = await api.cancelAllExecutions('running');
      addToast(`已停止 ${result.count} 个运行中的任务`, 'success');
      fetchExecutions();
    } catch (error) {
      console.error('Failed to cancel all:', error);
      addToast('停止任务失败', 'error');
    }
  };

  const handleRetry = async (id: string) => {
    setRetryingId(id);
    try {
      await api.retryExecution(id);
      addToast('任务已重新执行', 'success');
      fetchExecutions();
    } catch (error) {
      console.error('Failed to retry execution:', error);
      addToast('重试失败', 'error');
    } finally {
      setRetryingId(null);
    }
  };

  const handleViewDetails = (id: string) => {
    setSelectedExecutionId(id);
  };

  const getStatusIcon = (status: ExecutionStatus) => {
    switch (status) {
      case 'running':
        return '▶';
      case 'completed':
        return '✓';
      case 'failed':
        return '✗';
      case 'paused':
        return '⏸';
      case 'cancelled':
        return '⏹';
      default:
        return '○';
    }
  };

  const getStatusColor = (status: ExecutionStatus) => {
    switch (status) {
      case 'running':
        return '#1890ff';
      case 'completed':
        return '#52c41a';
      case 'failed':
        return '#f5222d';
      case 'paused':
        return '#faad14';
      default:
        return '#8c8c8c';
    }
  };

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${Math.round(ms / 1000)}s`;
    return `${Math.round(ms / 60000)}m`;
  };

  return (
    <div className="execution-list">
      <div className="header">
        <h2>任务列表</h2>
        <div className="actions">
          <select
            value={selectedStatus}
            onChange={(e) => setSelectedStatus(e.target.value as ExecutionStatus | '')}
          >
            <option value="">全部状态</option>
            <option value="pending">等待中</option>
            <option value="running">运行中</option>
            <option value="completed">已完成</option>
            <option value="failed">失败</option>
            <option value="paused">已暂停</option>
          </select>
          <button onClick={fetchExecutions}>刷新</button>
          <button onClick={handleCancelAll} className="danger">
            停止全部运行中
          </button>
        </div>
      </div>

      {loading ? (
        <div className="loading">加载中...</div>
      ) : (
        <div className="list">
          {executions.map((exec) => (
            <div key={exec.id} className="execution-item">
              <div className="status-icon" style={{ color: getStatusColor(exec.status) }}>
                {getStatusIcon(exec.status)}
              </div>

              <div className="info">
                <div className="title">
                  <span className="id">{exec.id}</span>
                  <span className="name">{exec.workflow_name}</span>
                </div>

                <div className="progress-bar">
                  <div
                    className="progress-fill"
                    style={{ width: `${exec.progress}%` }}
                  />
                </div>

                <div className="meta">
                  <span className="progress">{exec.progress}% {exec.status}</span>
                  {exec.current_step && (
                    <span>当前: {exec.current_step}</span>
                  )}
                  <span>
                    已完成: {exec.completed_steps}/{exec.total_steps} 步骤
                  </span>
                  <span>耗时: {formatDuration(exec.duration)}</span>
                </div>

                {exec.error && (
                  <div className="error">错误: {exec.error}</div>
                )}
              </div>

              <div className="actions">
                {exec.status === 'running' && (
                  <>
                    <button onClick={() => handlePause(exec.id)}>暂停</button>
                    <button onClick={() => handleCancel(exec.id)} className="danger">
                      停止
                    </button>
                  </>
                )}
                {exec.status === 'paused' && (
                  <>
                    <button onClick={() => handleResume(exec.id)}>恢复</button>
                    <button onClick={() => handleCancel(exec.id)} className="danger">
                      停止
                    </button>
                  </>
                )}
                {exec.status === 'failed' && (
                  <button
                    onClick={() => handleRetry(exec.id)}
                    disabled={retryingId === exec.id}
                  >
                    {retryingId === exec.id ? '重试中...' : '重试'}
                  </button>
                )}
                <button onClick={() => handleViewDetails(exec.id)}>详情</button>
              </div>
            </div>
          ))}

          {executions.length === 0 && (
            <div className="empty">暂无执行记录</div>
          )}
        </div>
      )}

      <div className="pagination">
        <button
          disabled={page === 1}
          onClick={() => setPage(page - 1)}
        >
          上一页
        </button>
        <span>
          第 {page} 页，共 {Math.ceil(total / 20)} 页，总计 {total} 条
        </span>
        <button
          disabled={page * 20 >= total}
          onClick={() => setPage(page + 1)}
        >
          下一页
        </button>
      </div>

      {selectedExecutionId && (
        <ExecutionDetailModal
          executionId={selectedExecutionId}
          onClose={() => setSelectedExecutionId(null)}
        />
      )}
    </div>
  );
};

export default ExecutionList;