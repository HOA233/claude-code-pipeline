import React, { useState, useCallback } from 'react';
import api from '../api/client';
import { useToast } from './Toast';
import './BatchExecution.css';

interface BatchItem {
  workflow_id: string;
  workflow_name: string;
  input?: Record<string, any>;
}

interface BatchExecutionProps {
  onComplete?: () => void;
}

export const BatchExecution: React.FC<BatchExecutionProps> = ({ onComplete }) => {
  const [batchItems, setBatchItems] = useState<BatchItem[]>([]);
  const [executing, setExecuting] = useState(false);
  const [results, setResults] = useState<any[]>([]);
  const { addToast } = useToast();

  const addBatchItem = useCallback((workflowId: string, workflowName: string, input?: Record<string, any>) => {
    setBatchItems((prev) => [
      ...prev,
      { workflow_id: workflowId, workflow_name: workflowName, input },
    ]);
  }, []);

  const removeBatchItem = useCallback((index: number) => {
    setBatchItems((prev) => prev.filter((_, i) => i !== index));
  }, []);

  const executeBatch = useCallback(async () => {
    if (batchItems.length === 0) {
      addToast('请先添加要执行的工作流', 'warning');
      return;
    }

    setExecuting(true);
    setResults([]);

    try {
      const executionPromises = batchItems.map(async (item) => {
        try {
          const execution = await api.executeWorkflow({
            workflow_id: item.workflow_id,
            input: item.input,
            async: true,
          });
          return {
            workflow_name: item.workflow_name,
            execution_id: execution.id,
            status: 'started',
          };
        } catch (error: any) {
          return {
            workflow_name: item.workflow_name,
            error: error.message || '执行失败',
            status: 'failed',
          };
        }
      });

      const batchResults = await Promise.all(executionPromises);
      setResults(batchResults);

      const successCount = batchResults.filter((r) => r.status === 'started').length;
      addToast(`批量执行已启动: ${successCount}/${batchItems.length} 个工作流`, 'success');

      if (onComplete) {
        onComplete();
      }
    } catch (error) {
      console.error('Batch execution failed:', error);
      addToast('批量执行失败', 'error');
    } finally {
      setExecuting(false);
    }
  }, [batchItems, addToast, onComplete]);

  const clearBatch = useCallback(() => {
    setBatchItems([]);
    setResults([]);
  }, []);

  return (
    <div className="batch-execution">
      <div className="batch-header">
        <h3>批量执行</h3>
        <div className="batch-actions">
          <button
            className="btn-secondary"
            onClick={clearBatch}
            disabled={batchItems.length === 0 || executing}
          >
            清空
          </button>
          <button
            className="btn-primary"
            onClick={executeBatch}
            disabled={batchItems.length === 0 || executing}
          >
            {executing ? '执行中...' : `执行全部 (${batchItems.length})`}
          </button>
        </div>
      </div>

      <div className="batch-list">
        {batchItems.length === 0 ? (
          <div className="batch-empty">
            暂无批量执行任务。请从工作流列表添加要执行的工作流。
          </div>
        ) : (
          batchItems.map((item, index) => (
            <div key={index} className="batch-item">
              <div className="batch-item-info">
                <span className="batch-item-name">{item.workflow_name}</span>
                <span className="batch-item-id">{item.workflow_id}</span>
              </div>
              <button
                className="batch-item-remove"
                onClick={() => removeBatchItem(index)}
                disabled={executing}
              >
                ×
              </button>
            </div>
          ))
        )}
      </div>

      {results.length > 0 && (
        <div className="batch-results">
          <h4>执行结果</h4>
          <div className="results-list">
            {results.map((result, index) => (
              <div key={index} className={`result-item result-${result.status}`}>
                <span className="result-name">{result.workflow_name}</span>
                {result.execution_id && (
                  <span className="result-id">{result.execution_id}</span>
                )}
                {result.error && (
                  <span className="result-error">{result.error}</span>
                )}
                <span className="result-status">
                  {result.status === 'started' ? '✓ 已启动' : '✗ 失败'}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export default BatchExecution;