import React, { useState, useEffect, useCallback } from 'react';
import type { Agent } from '../types';
import api from '../api/client';

interface AgentListProps {
  tenantId?: string;
  onSelect?: (agent: Agent) => void;
}

export const AgentList: React.FC<AgentListProps> = ({ tenantId, onSelect }) => {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState('');

  const fetchAgents = useCallback(async () => {
    setLoading(true);
    try {
      const result = await api.listAgents({
        tenant_id: tenantId,
        category: selectedCategory || undefined,
      });
      setAgents(result.agents);
    } catch (error) {
      console.error('Failed to fetch agents:', error);
    } finally {
      setLoading(false);
    }
  }, [tenantId, selectedCategory]);

  useEffect(() => {
    fetchAgents();
  }, [fetchAgents]);

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
      await api.updateAgent(agent.id, { enabled: !agent.enabled });
      fetchAgents();
    } catch (error) {
      console.error('Failed to toggle agent:', error);
    }
  };

  const getCategoryColor = (category: string) => {
    const colors: Record<string, string> = {
      analysis: '#1890ff',
      testing: '#52c41a',
      security: '#f5222d',
      performance: '#fa8c16',
      documentation: '#722ed1',
      refactoring: '#13c2c2',
      debugging: '#eb2f96',
    };
    return colors[category] || '#8c8c8c';
  };

  const categories = [...new Set(agents.map((a) => a.category).filter(Boolean))];

  return (
    <div className="agent-list">
      <div className="header">
        <h2>Agent 列表</h2>
        <div className="actions">
          <select
            value={selectedCategory}
            onChange={(e) => setSelectedCategory(e.target.value)}
          >
            <option value="">全部分类</option>
            {categories.map((cat) => (
              <option key={cat} value={cat}>
                {cat}
              </option>
            ))}
          </select>
          <button onClick={fetchAgents}>刷新</button>
        </div>
      </div>

      {loading ? (
        <div className="loading">加载中...</div>
      ) : (
        <div className="grid">
          {agents.map((agent) => (
            <div
              key={agent.id}
              className={`agent-card ${!agent.enabled ? 'disabled' : ''}`}
              onClick={() => onSelect?.(agent)}
            >
              <div className="card-header">
                <span
                  className="category-badge"
                  style={{ backgroundColor: getCategoryColor(agent.category) }}
                >
                  {agent.category || 'general'}
                </span>
                <span className={`status ${agent.enabled ? 'enabled' : 'disabled'}`}>
                  {agent.enabled ? '启用' : '禁用'}
                </span>
              </div>

              <h3>{agent.name}</h3>
              <p className="description">{agent.description}</p>

              <div className="meta">
                <span className="model">{agent.model}</span>
                {agent.max_tokens > 0 && (
                  <span className="tokens">{agent.max_tokens} tokens</span>
                )}
              </div>

              {agent.tags && agent.tags.length > 0 && (
                <div className="tags">
                  {agent.tags.map((tag) => (
                    <span key={tag} className="tag">
                      {tag}
                    </span>
                  ))}
                </div>
              )}

              <div className="card-actions">
                <button onClick={(e) => { e.stopPropagation(); handleToggle(agent); }}>
                  {agent.enabled ? '禁用' : '启用'}
                </button>
                <button onClick={(e) => { e.stopPropagation(); handleDelete(agent.id); }}>
                  删除
                </button>
              </div>
            </div>
          ))}

          {agents.length === 0 && (
            <div className="empty">暂无 Agent</div>
          )}
        </div>
      )}
    </div>
  );
};

export default AgentList;