import React, { useState, useEffect } from 'react'
import { WorkflowList } from '../components/WorkflowList'
import api from '../api/client'
import './Workflows.css'

function Workflows() {
  const [activeTab, setActiveTab] = useState('workflows')
  const [templates, setTemplates] = useState([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (activeTab === 'templates') {
      fetchTemplates()
    }
  }, [activeTab])

  const fetchTemplates = async () => {
    setLoading(true)
    try {
      const [builtin, custom] = await Promise.all([
        api.getBuiltinTemplates(),
        api.getCustomTemplates(),
      ])
      setTemplates([...(builtin.templates || []), ...(custom.templates || [])])
    } catch (error) {
      console.error('Failed to fetch templates:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleUseTemplate = async (template) => {
    const name = prompt('请输入工作流名称:', template.name)
    if (!name) return

    try {
      const workflow = await api.instantiateTemplate(template.id, name)
      alert(`工作流 "${name}" 已创建`)
      setActiveTab('workflows')
    } catch (error) {
      console.error('Failed to instantiate template:', error)
      alert('创建失败')
    }
  }

  return (
    <div className="workflows-page">
      <div className="page-header">
        <h1>工作流管理</h1>
        <p>配置和组合多个 Agent 形成工作流</p>
      </div>

      <div className="tabs-bar">
        <button
          className={`tab-btn ${activeTab === 'workflows' ? 'active' : ''}`}
          onClick={() => setActiveTab('workflows')}
        >
          我的工作流
        </button>
        <button
          className={`tab-btn ${activeTab === 'templates' ? 'active' : ''}`}
          onClick={() => setActiveTab('templates')}
        >
          模板库
        </button>
      </div>

      {activeTab === 'workflows' ? (
        <WorkflowList />
      ) : (
        <div className="templates-section">
          {loading ? (
            <div className="loading">加载中...</div>
          ) : (
            <div className="templates-grid">
              {templates.map((template) => (
                <div key={template.id} className="template-card">
                  <div className="template-icon">{template.icon || '📋'}</div>
                  <div className="template-content">
                    <h3>{template.name}</h3>
                    <p>{template.description}</p>
                    <div className="template-meta">
                      <span className="category">{template.category}</span>
                      <span className="agents">{template.agents?.length || 0} 个 Agent</span>
                    </div>
                  </div>
                  <button
                    className="btn-use-template"
                    onClick={() => handleUseTemplate(template)}
                  >
                    使用模板
                  </button>
                </div>
              ))}
              {templates.length === 0 && (
                <div className="empty">暂无模板</div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default Workflows