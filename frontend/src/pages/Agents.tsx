import React, { useState, useEffect, useCallback } from 'react'
import type { Agent } from '../types'
import api from '../api/client'
import { AgentForm } from '../components/AgentForm'
import './Agents.css'

function Agents() {
  const [agents, setAgents] = useState<Agent[]>([])
  const [loading, setLoading] = useState(false)
  const [showForm, setShowForm] = useState(false)
  const [editingAgent, setEditingAgent] = useState<Agent | null>(null)
  const [selectedCategory, setSelectedCategory] = useState('')

  const fetchAgents = useCallback(async () => {
    setLoading(true)
    try {
      const result = await api.listAgents({ category: selectedCategory || undefined })
      setAgents(result.agents)
    } catch (error) {
      console.error('Failed to fetch agents:', error)
    } finally {
      setLoading(false)
    }
  }, [selectedCategory])

  useEffect(() => {
    fetchAgents()
  }, [fetchAgents])

  const handleCreate = () => {
    setEditingAgent(null)
    setShowForm(true)
  }

  const handleEdit = (agent) => {
    setEditingAgent(agent)
    setShowForm(true)
  }

  const handleDelete = async (id) => {
    if (!confirm('确定要删除此 Agent 吗？')) return
    try {
      await api.deleteAgent(id)
      fetchAgents()
    } catch (error) {
      console.error('Failed to delete agent:', error)
    }
  }

  const handleToggle = async (agent) => {
    try {
      await api.updateAgent(agent.id, { enabled: !agent.enabled })
      fetchAgents()
    } catch (error) {
      console.error('Failed to toggle agent:', error)
    }
  }

  const handleExecute = async (agent) => {
    const inputStr = prompt('请输入执行参数 (JSON 格式):', '{}')
    if (!inputStr) return
    try {
      const input = JSON.parse(inputStr)
      const execution = await api.executeAgent(agent.id, { input, async: true })
      alert(`执行已创建: ${execution.id}`)
    } catch (error) {
      console.error('Failed to execute agent:', error)
      alert('执行失败: ' + error.message)
    }
  }

  const handleFormSuccess = () => {
    setShowForm(false)
    setEditingAgent(null)
    fetchAgents()
  }

  const getCategoryColor = (category) => {
    const colors = {
      'code-review': '#52c41a',
      'testing': '#1890ff',
      'security': '#f5222d',
      'documentation': '#722ed1',
      'refactoring': '#faad14',
      'general': '#8c8c8c',
    }
    return colors[category || 'general'] || '#8c8c8c'
  }

  const categories = [...new Set(agents.map(a => a.category).filter(Boolean))]

  if (showForm) {
    return (
      <div className="agents-page">
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
            <div className="modal-body" style={{ maxHeight: '70vh', overflowY: 'auto' }}>
              <AgentForm
                agent={editingAgent || undefined}
                onSuccess={handleFormSuccess}
                onCancel={() => setShowForm(false)}
              />
            </div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="agents-page">
      <div className="page-header">
        <h1>Agent 管理</h1>
        <p>管理 Claude Code Agent 配置</p>
      </div>

      <div className="agents-toolbar">
        <div className="filters">
          <select
            value={selectedCategory}
            onChange={(e) => setSelectedCategory(e.target.value)}
          >
            <option value="">全部分类</option>
            {categories.map((cat) => (
              <option key={cat} value={cat}>{cat}</option>
            ))}
          </select>
          <button onClick={fetchAgents}>刷新</button>
        </div>
        <button className="btn-primary" onClick={handleCreate}>
          + 创建 Agent
        </button>
      </div>

      {loading ? (
        <div className="loading">加载中...</div>
      ) : (
        <div className="agents-grid">
          {agents.map((agent) => (
            <div key={agent.id} className={`agent-card ${!agent.enabled ? 'disabled' : ''}`}>
              <div className="card-header">
                <span
                  className="category-badge"
                  style={{ backgroundColor: getCategoryColor(agent.category) }}
                >
                  {agent.category || 'general'}
                </span>
                <span className={`status-badge ${agent.enabled ? 'enabled' : 'disabled'}`}>
                  {agent.enabled ? '启用' : '禁用'}
                </span>
              </div>

              <h3>{agent.name}</h3>
              <p className="description">{agent.description}</p>

              <div className="meta">
                <span className="model">{agent.model}</span>
                <span className="timeout">超时: {agent.timeout}s</span>
              </div>

              {agent.tags && agent.tags.length > 0 && (
                <div className="tags">
                  {agent.tags.slice(0, 3).map((tag) => (
                    <span key={tag} className="tag">{tag}</span>
                  ))}
                  {agent.tags.length > 3 && (
                    <span className="tag">+{agent.tags.length - 3}</span>
                  )}
                </div>
              )}

              <div className="card-actions">
                <button className="btn-execute" onClick={() => handleExecute(agent)}>
                  执行
                </button>
                <button onClick={() => handleToggle(agent)}>
                  {agent.enabled ? '禁用' : '启用'}
                </button>
                <button onClick={() => handleEdit(agent)}>编辑</button>
                <button className="btn-danger" onClick={() => handleDelete(agent.id)}>
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
  )
}

export default Agents