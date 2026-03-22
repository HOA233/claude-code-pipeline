import React, { useState } from 'react';
import api from '../api/client';
import type { Workflow, WorkflowCreateRequest, AgentNode, Agent } from '../types';

interface WorkflowFormProps {
  workflow?: Workflow;
  agents: Agent[];
  onSuccess?: (workflow: Workflow) => void;
  onCancel?: () => void;
}

export const WorkflowForm: React.FC<WorkflowFormProps> = ({
  workflow,
  agents,
  onSuccess,
  onCancel,
}) => {
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState<WorkflowCreateRequest>({
    name: workflow?.name || '',
    description: workflow?.description || '',
    mode: workflow?.mode || 'serial',
    agents: workflow?.agents || [],
    context: workflow?.context || {},
  });

  const [newNode, setNewNode] = useState<Partial<AgentNode>>({
    agent_id: '',
    input: {},
    depends_on: [],
  });

  const addAgentNode = () => {
    if (newNode.agent_id) {
      const agent = agents.find((a) => a.id === newNode.agent_id);
      const node: AgentNode = {
        id: `node-${Date.now()}`,
        agent_id: newNode.agent_id || '',
        name: agent?.name || '',
        input: newNode.input || {},
        depends_on: newNode.depends_on || [],
      };
      setFormData({
        ...formData,
        agents: [...formData.agents, node],
      });
      setNewNode({ agent_id: '', input: {}, depends_on: [] });
    }
  };

  const removeAgentNode = (nodeId: string) => {
    setFormData({
      ...formData,
      agents: formData.agents.filter((n) => n.id !== nodeId),
    });
  };

  const updateNodeDependsOn = (nodeId: string, dependsOn: string[]) => {
    setFormData({
      ...formData,
      agents: formData.agents.map((n) =>
        n.id === nodeId ? { ...n, depends_on: dependsOn } : n
      ),
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      let result: Workflow;
      if (workflow?.id) {
        result = await api.updateWorkflow(workflow.id, formData);
      } else {
        result = await api.createWorkflow(formData);
      }
      onSuccess?.(result);
    } catch (error) {
      console.error('Failed to save workflow:', error);
      alert('保存失败');
    } finally {
      setLoading(false);
    }
  };

  const getAvailableDependencies = (currentNodeId: string) => {
    return formData.agents.filter((n) => n.id !== currentNodeId);
  };

  return (
    <form className="workflow-form" onSubmit={handleSubmit}>
      <div className="form-section">
        <h3>基本信息</h3>
        <div className="form-group">
          <label>名称 *</label>
          <input
            type="text"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            required
            placeholder="工作流名称"
          />
        </div>

        <div className="form-group">
          <label>描述</label>
          <textarea
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            placeholder="工作流功能描述"
            rows={3}
          />
        </div>

        <div className="form-group">
          <label>执行模式</label>
          <select
            value={formData.mode}
            onChange={(e) =>
              setFormData({ ...formData, mode: e.target.value as any })
            }
          >
            <option value="serial">串行执行</option>
            <option value="parallel">并行执行</option>
            <option value="hybrid">混合模式</option>
          </select>
        </div>
      </div>

      <div className="form-section">
        <h3>Agent 节点</h3>
        <div className="node-add-form">
          <select
            value={newNode.agent_id}
            onChange={(e) => setNewNode({ ...newNode, agent_id: e.target.value })}
          >
            <option value="">选择 Agent</option>
            {agents.map((agent) => (
              <option key={agent.id} value={agent.id}>
                {agent.name}
              </option>
            ))}
          </select>
          <button type="button" onClick={addAgentNode} className="btn-add">
            添加节点
          </button>
        </div>

        <div className="node-list">
          {formData.agents.map((node, index) => {
            const agent = agents.find((a) => a.id === node.agent_id);
            const availableDeps = getAvailableDependencies(node.id);

            return (
              <div key={node.id} className="node-item">
                <div className="node-header">
                  <span className="node-index">{index + 1}</span>
                  <span className="node-name">{agent?.name || node.agent_id}</span>
                  <button
                    type="button"
                    onClick={() => removeAgentNode(node.id)}
                    className="btn-remove"
                  >
                    ×
                  </button>
                </div>
                <div className="node-config">
                  <label>依赖节点:</label>
                  <div className="dep-checkboxes">
                    {availableDeps.map((dep) => {
                      const depAgent = agents.find((a) => a.id === dep.agent_id);
                      return (
                        <label key={dep.id} className="checkbox-label">
                          <input
                            type="checkbox"
                            checked={node.depends_on?.includes(dep.id) || false}
                            onChange={(e) => {
                              const deps = node.depends_on || [];
                              const newDeps = e.target.checked
                                ? [...deps, dep.id]
                                : deps.filter((d) => d !== dep.id);
                              updateNodeDependsOn(node.id, newDeps);
                            }}
                          />
                          {depAgent?.name || dep.id}
                        </label>
                      );
                    })}
                    {availableDeps.length === 0 && (
                      <span className="no-deps">无可用依赖</span>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
          {formData.agents.length === 0 && (
            <div className="empty-nodes">暂无节点，请添加 Agent</div>
          )}
        </div>
      </div>

      <div className="form-section">
        <h3>执行预览</h3>
        <div className="workflow-preview">
          <div className="preview-diagram">
            {formData.mode === 'serial' && (
              <div className="diagram serial">
                {formData.agents.map((node, i) => {
                  const agent = agents.find((a) => a.id === node.agent_id);
                  return (
                    <React.Fragment key={node.id}>
                      <div className="preview-node">{agent?.name || node.agent_id}</div>
                      {i < formData.agents.length - 1 && (
                        <div className="preview-arrow">→</div>
                      )}
                    </React.Fragment>
                  );
                })}
              </div>
            )}
            {formData.mode === 'parallel' && (
              <div className="diagram parallel">
                {formData.agents.map((node) => {
                  const agent = agents.find((a) => a.id === node.agent_id);
                  return (
                    <div key={node.id} className="preview-node">
                      {agent?.name || node.agent_id}
                    </div>
                  );
                })}
              </div>
            )}
            {formData.mode === 'hybrid' && (
              <div className="diagram hybrid">
                {formData.agents.map((node) => {
                  const agent = agents.find((a) => a.id === node.agent_id);
                  const hasDeps = (node.depends_on?.length || 0) > 0;
                  return (
                    <div
                      key={node.id}
                      className={`preview-node ${hasDeps ? 'with-deps' : ''}`}
                    >
                      {agent?.name || node.agent_id}
                      {hasDeps && (
                        <div className="deps-indicator">
                          ↑ {(node.depends_on || []).length} 依赖
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>
      </div>

      <div className="form-actions">
        <button type="button" onClick={onCancel} className="btn-secondary">
          取消
        </button>
        <button type="submit" disabled={loading} className="btn-primary">
          {loading ? '保存中...' : workflow ? '更新' : '创建'}
        </button>
      </div>
    </form>
  );
};

export default WorkflowForm;