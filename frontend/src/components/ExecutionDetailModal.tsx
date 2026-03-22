import React, { useState, useEffect } from 'react';
import api from '../api/client';
import type { Execution, NodeResult } from '../types';
import './ExecutionDetailModal.css';

interface ExecutionDetailModalProps {
  executionId: string;
  onClose: () => void;
}

export const ExecutionDetailModal: React.FC<ExecutionDetailModalProps> = ({
  executionId,
  onClose,
}) => {
  const [execution, setExecution] = useState<Execution | null>(null);
  const [logs, setLogs] = useState<any[]>([]);
  const [activeTab, setActiveTab] = useState<'details' | 'logs' | 'nodes'>('details');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const [execData, logsData] = await Promise.all([
          api.getExecution(executionId),
          api.getExecutionLogs(executionId, 100),
        ]);
        setExecution(execData);
        setLogs(logsData.logs || []);
      } catch (error) {
        console.error('Failed to fetch execution details:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [executionId]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return '#52c41a';
      case 'failed': return '#f5222d';
      case 'running': return '#1890ff';
      case 'paused': return '#faad14';
      default: return '#8c8c8c';
    }
  };

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${Math.round(ms / 1000)}s`;
    return `${Math.round(ms / 60000)}m`;
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('zh-CN');
  };

  if (loading) {
    return (
      <div className="modal-overlay">
        <div className="modal-content execution-detail-modal">
          <div className="modal-loading">加载中...</div>
        </div>
      </div>
    );
  }

  if (!execution) {
    return (
      <div className="modal-overlay">
        <div className="modal-content execution-detail-modal">
          <div className="modal-error">执行不存在</div>
          <button onClick={onClose}>关闭</button>
        </div>
      </div>
    );
  }

  const nodeResults = Object.entries(execution.node_results || {});

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content execution-detail-modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <div className="execution-title">
            <h3>{execution.workflow_name}</h3>
            <span
              className="execution-status"
              style={{ backgroundColor: getStatusColor(execution.status) }}
            >
              {execution.status}
            </span>
          </div>
          <button className="modal-close" onClick={onClose}>×</button>
        </div>

        <div className="execution-summary">
          <div className="summary-item">
            <span className="label">执行 ID</span>
            <span className="value mono">{execution.id}</span>
          </div>
          <div className="summary-item">
            <span className="label">进度</span>
            <span className="value">{execution.progress}%</span>
          </div>
          <div className="summary-item">
            <span className="label">耗时</span>
            <span className="value">{formatDuration(execution.duration)}</span>
          </div>
          <div className="summary-item">
            <span className="label">步骤</span>
            <span className="value">{execution.completed_steps}/{execution.total_steps}</span>
          </div>
          <div className="summary-item">
            <span className="label">创建时间</span>
            <span className="value">{formatDate(execution.created_at)}</span>
          </div>
        </div>

        <div className="tab-bar">
          <button
            className={`tab ${activeTab === 'details' ? 'active' : ''}`}
            onClick={() => setActiveTab('details')}
          >
            详情
          </button>
          <button
            className={`tab ${activeTab === 'nodes' ? 'active' : ''}`}
            onClick={() => setActiveTab('nodes')}
          >
            节点结果 ({nodeResults.length})
          </button>
          <button
            className={`tab ${activeTab === 'logs' ? 'active' : ''}`}
            onClick={() => setActiveTab('logs')}
          >
            日志 ({logs.length})
          </button>
        </div>

        <div className="modal-body">
          {activeTab === 'details' && (
            <div className="details-tab">
              {execution.error && (
                <div className="error-box">
                  <strong>错误信息:</strong>
                  <pre>{execution.error}</pre>
                </div>
              )}

              <div className="progress-section">
                <h4>执行进度</h4>
                <div className="progress-bar-large">
                  <div
                    className="progress-fill"
                    style={{ width: `${execution.progress}%` }}
                  />
                </div>
                <div className="progress-info">
                  <span>{execution.current_step || '等待中...'}</span>
                  <span>{execution.progress}%</span>
                </div>
              </div>

              {execution.final_output && (
                <div className="output-section">
                  <h4>执行结果</h4>
                  <pre className="output-code">
                    {JSON.stringify(execution.final_output, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          )}

          {activeTab === 'nodes' && (
            <div className="nodes-tab">
              {nodeResults.length > 0 ? (
                <div className="node-results">
                  {nodeResults.map(([nodeId, result]) => (
                    <div key={nodeId} className="node-result-card">
                      <div className="node-result-header">
                        <span className="node-id">{nodeId}</span>
                        <span
                          className="node-status"
                          style={{ color: getStatusColor(result.status) }}
                        >
                          {result.status}
                        </span>
                      </div>
                      <div className="node-result-meta">
                        <span>Agent: {result.agent_id}</span>
                        <span>耗时: {formatDuration(result.duration)}</span>
                        <span>重试: {result.retries}</span>
                      </div>
                      {result.error && (
                        <div className="node-error">{result.error}</div>
                      )}
                      {result.output && (
                        <details className="node-output">
                          <summary>输出结果</summary>
                          <pre>{JSON.stringify(result.output, null, 2)}</pre>
                        </details>
                      )}
                    </div>
                  ))}
                </div>
              ) : (
                <div className="empty-tab">暂无节点结果</div>
              )}
            </div>
          )}

          {activeTab === 'logs' && (
            <div className="logs-tab">
              {logs.length > 0 ? (
                <div className="log-list">
                  {logs.map((log, index) => (
                    <div key={index} className={`log-entry log-${log.level}`}>
                      <span className="log-time">
                        {new Date(log.timestamp).toLocaleTimeString()}
                      </span>
                      <span className="log-level">{log.level.toUpperCase()}</span>
                      <span className="log-message">{log.message}</span>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="empty-tab">暂无日志</div>
              )}
            </div>
          )}
        </div>

        <div className="modal-footer">
          <button onClick={onClose} className="btn-secondary">关闭</button>
          {execution.status === 'failed' && (
            <button
              onClick={async () => {
                try {
                  await api.retryExecution(executionId);
                  onClose();
                } catch (error) {
                  console.error('Failed to retry:', error);
                }
              }}
              className="btn-primary"
            >
              重试
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default ExecutionDetailModal;