import React, { useState, useEffect, useCallback } from 'react';
import type { Workflow, ExecutionMode } from '../types';
import api from '../api/client';

interface WorkflowListProps {
  tenantId?: string;
  onSelect?: (workflow: Workflow) => void;
  onExecute?: (workflow: Workflow) => void;
}

export const WorkflowList: React.FC<WorkflowListProps> = ({
  tenantId,
  onSelect,
  onExecute,
}) => {
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchWorkflows = useCallback(async () => {
    setLoading(true);
    try {
      const result = await api.listWorkflows({ tenant_id: tenantId });
      setWorkflows(result.workflows);
    } catch (error) {
      console.error('Failed to fetch workflows:', error);
    } finally {
      setLoading(false);
    }
  }, [tenantId]);

  useEffect(() => {
    fetchWorkflows();
  }, [fetchWorkflows]);

  const handleDelete = async (id: string) => {
    if (!confirm('确定要删除此工作流吗？')) return;
    try {
      await api.deleteWorkflow(id);
      fetchWorkflows();
    } catch (error) {
      console.error('Failed to delete workflow:', error);
    }
  };

  const handleToggle = async (workflow: Workflow) => {
    try {
      const enabled = !workflow.enabled;
      await fetch(`/api/workflows/${workflow.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ enabled }),
      });
      fetchWorkflows();
    } catch (error) {
      console.error('Failed to toggle workflow:', error);
    }
  };

  const getModeIcon = (mode: ExecutionMode) => {
    switch (mode) {
      case 'serial':
        return '→';
      case 'parallel':
        return '⇉';
      case 'hybrid':
        return '⇄';
      default:
        return '○';
    }
  };

  const getModeLabel = (mode: ExecutionMode) => {
    switch (mode) {
      case 'serial':
        return '串行';
      case 'parallel':
        return '并行';
      case 'hybrid':
        return '混合';
      default:
        return mode;
    }
  };

  return (
    <div className="workflow-list">
      <div className="header">
        <h2>工作流列表</h2>
        <button onClick={fetchWorkflows}>刷新</button>
      </div>

      {loading ? (
        <div className="loading">加载中...</div>
      ) : (
        <div className="list">
          {workflows.map((workflow) => (
            <div
              key={workflow.id}
              className={`workflow-item ${!workflow.enabled ? 'disabled' : ''}`}
              onClick={() => onSelect?.(workflow)}
            >
              <div className="main-info">
                <div className="title-row">
                  <span className="mode-icon" title={getModeLabel(workflow.mode)}>
                    {getModeIcon(workflow.mode)}
                  </span>
                  <h3>{workflow.name}</h3>
                  <span className={`status ${workflow.enabled ? 'enabled' : 'disabled'}`}>
                    {workflow.enabled ? '启用' : '禁用'}
                  </span>
                </div>

                <p className="description">{workflow.description}</p>

                <div className="meta">
                  <span className="agents-count">
                    {workflow.agents?.length || 0} 个 Agent
                  </span>
                  <span className="mode">{getModeLabel(workflow.mode)}</span>
                </div>

                {workflow.agents && workflow.agents.length > 0 && (
                  <div className="agents-preview">
                    {workflow.agents.slice(0, 3).map((agent) => (
                      <span key={agent.id} className="agent-node">
                        {agent.name || agent.id}
                      </span>
                    ))}
                    {workflow.agents.length > 3 && (
                      <span className="more">+{workflow.agents.length - 3}</span>
                    )}
                  </div>
                )}
              </div>

              <div className="actions">
                {workflow.enabled && (
                  <button
                    className="execute"
                    onClick={(e) => {
                      e.stopPropagation();
                      onExecute?.(workflow);
                    }}
                  >
                    执行
                  </button>
                )}
                <button onClick={(e) => { e.stopPropagation(); handleToggle(workflow); }}>
                  {workflow.enabled ? '禁用' : '启用'}
                </button>
                <button
                  className="delete"
                  onClick={(e) => { e.stopPropagation(); handleDelete(workflow.id); }}
                >
                  删除
                </button>
              </div>
            </div>
          ))}

          {workflows.length === 0 && (
            <div className="empty">暂无工作流</div>
          )}
        </div>
      )}
    </div>
  );
};

export default WorkflowList;