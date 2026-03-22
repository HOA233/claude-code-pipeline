import React, { useState, useEffect, useCallback } from 'react';
import type { Agent } from '../types';
import api from '../api/client';
import { AgentForm } from './AgentForm';
import './FormComponents.css';

interface AgentListProps {
  tenantId?: string;
}

export const AgentList: React.FC<AgentListProps> = ({ tenantId }) => {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(false);
  const [showForm, setShowForm] = useState(false);
  const [editingAgent, setEditingAgent] = useState<Agent | null>(null);

  const fetchAgents = useCallback(async () => {
    setLoading(true);
    try {
      const result = await api.listAgents({ tenant_id: tenantId });
      setAgents(result.agents);
    } catch (error) {
      console.error('Failed to fetch agents:', error);
    } finally {
      setLoading(false);
    }
  }, [tenantId]);

  useEffect(() => {
    fetchAgents();
  }, [fetchAgents]);

  const handleCreate = () => {
    setEditingAgent(null);
    setShowForm(true);
  };

  const handleEdit = (agent: Agent) => {
    setEditingAgent(agent);
    setShowForm(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('确定要删除此 Agent 吗？')) return;
    try {
      await api.deleteAgent(id);
      fetchAgents();
    } catch (error) {
      console.error('Failed to delete agent:', error);
    }
  };

  const handleToggle = async (agent: Agent) => {
    try {
      if (agent.enabled) {
        // Disable - update the agent
        await api.updateAgent(agent.id, { enabled: false });
      } else {
        await api.updateAgent(agent.id, { enabled: true });
      }
      fetchAgents();
    } catch (error) {
      console.error('Failed to toggle agent:', error);
    }
  };

  const handleExecute = async (agent: Agent) => {
    const inputStr = prompt('请输入执行参数 (JSON 格式):', '{}');
    if (!inputStr) return;

    try {
      const input = JSON.parse(inputStr);
      const execution = await api.executeAgent(agent.id, { input, async: true });
      alert(`执行已创建: ${execution.id}`);
    } catch (error) {
      console.error('Failed to execute agent:', error);
      alert('执行失败');
    }
  };

  const handleFormSuccess = () => {
    setShowForm(false);
    setEditingAgent(null);
    fetchAgents();
  };

  const getCategoryColor = (category?: string) => {
    const colors: Record<string, string> = {
      'code-review': '#52c41a',
      'testing': '#1890ff',
      'security': '#f5222d',
      'documentation': '#722ed1',
      'refactoring': '#faad14',
      'general': '#8c8c8c',
    };
    return colors[category || 'general'] || '#8c8c8c';
  };

  if (showForm) {
    return (
      <div className="modal-overlay">
        <div className="modal-content" style={{ maxWidth: '800px' }}>
          <div className="modal-header">
            <h3 className="modal-title">
              {editingAgent ? '编辑 Agent' : '创建 Agent'}
            </h3>
            <button className="modal-close" onClick={() => setShowForm(false)}>
              ×
            </button>
          </div>
          <div className="modal-body">
            <AgentForm
              agent={editingAgent || undefined}
              onSuccess={handleFormSuccess}
              onCancel={() => setShowForm(false)}
            />
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="agent-list-container">
      <div className="list-header">
        <button className="btn-primary" onClick={handleCreate}>
          + 创建 Agent
        </button>
      </div>

      {loading ? (
        <div className="loading">加载中...</div>
      ) : (
        <div className="list">
          {agents.map((agent) => (
            <div key={agent.id} className={`agent-item ${!agent.enabled ? 'disabled' : ''}`}>
              <div className="agent-main">
                <div className="agent-header">
                  <h3>{agent.name}</h3>
                  <span
                    className={`status-badge ${agent.enabled ? 'enabled' : 'disabled'}`}
                  >
                    {agent.enabled ? '启用' : '禁用'}
                  </span>
                </div>

                <p className="agent-description">{agent.description}</p>

                <div className="agent-meta">
                  <span
                    className="category-badge"
                    style={{ backgroundColor: getCategoryColor(agent.category) }}
                  >
                    {agent.category || 'general'}
                  </span>
                  <span className="model-badge">{agent.model}</span>
                  {agent.tags?.map((tag) => (
                    <span key={tag} className="tag-badge">
                      {tag}
                    </span>
                  ))}
                </div>

                <div className="agent-stats">
                  <span>工具: {agent.tools?.length || 0}</span>
                  <span>权限: {agent.permissions?.length || 0}</span>
                  <span>超时: {agent.timeout}s</span>
                </div>
              </div>

              <div className="agent-actions">
                <button onClick={() => handleExecute(agent)} className="btn-execute">
                  执行
                </button>
                <button onClick={() => handleToggle(agent)}>
                  {agent.enabled ? '禁用' : '启用'}
                </button>
                <button onClick={() => handleEdit(agent)}>编辑</button>
                <button
                  className="btn-danger"
                  onClick={() => handleDelete(agent.id)}
                >
                  删除
                </button>
              </div>
            </div>
          ))}

          {agents.length === 0 && (
            <div className="empty-state">
              <div className="empty-state-icon">🤖</div>
              <div className="empty-state-text">暂无 Agent，点击上方按钮创建</div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default AgentList;