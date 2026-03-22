import React, { useState, useEffect, useCallback } from 'react'
import { WorkflowList } from '../components/WorkflowList'
import { BatchExecution } from '../components/BatchExecution'
import { useToast } from '../components/Toast'
import type { Workflow } from '../types'
import api from '../api/client'
import './Workflows.css'

function Workflows() {
  const [activeTab, setActiveTab] = useState('workflows')
  const [templates, setTemplates] = useState([])
  const [loading, setLoading] = useState(false)
  const [showBatchPanel, setShowBatchPanel] = useState(false)
  const [batchItems, setBatchItems] = useState<{ workflow_id: string; workflow_name: string }[]>([])
  const { addToast } = useToast()

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
      setTemplates([...(builtin.templates || builtin || []), ...(custom.templates || custom || [])])
    } catch (error) {
      console.error('Failed to fetch templates:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleUseTemplate = async (template: any) => {
    const name = prompt('请输入工作流名称:', template.name)
    if (!name) return

    try {
      await api.instantiateTemplate(template.id, name)
      addToast(`工作流 "${name}" 已创建`, 'success')
      setActiveTab('workflows')
    } catch (error) {
      console.error('Failed to instantiate template:', error)
      addToast('创建失败', 'error')
    }
  }

  const handleExecuteWorkflow = useCallback((workflow: Workflow) => {
    const execute = async () => {
      try {
        const execution = await api.executeWorkflow({
          workflow_id: workflow.id,
          async: true,
        })
        addToast(`工作流 "${workflow.name}" 已开始执行`, 'success')
      } catch (error) {
        console.error('Failed to execute workflow:', error)
        addToast('执行失败', 'error')
      }
    }
    execute()
  }, [addToast])

  const handleAddToBatch = useCallback((workflow: Workflow) => {
    setBatchItems((prev) => {
      if (prev.some((item) => item.workflow_id === workflow.id)) {
        addToast('该工作流已在批量执行列表中', 'warning')
        return prev
      }
      addToast(`已添加 "${workflow.name}" 到批量执行`, 'info')
      return [...prev, { workflow_id: workflow.id, workflow_name: workflow.name }]
    })
    setShowBatchPanel(true)
  }, [addToast])

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
        {batchItems.length > 0 && (
          <button
            className={`tab-btn batch-tab ${showBatchPanel ? 'active' : ''}`}
            onClick={() => setShowBatchPanel(!showBatchPanel)}
          >
            批量执行 ({batchItems.length})
          </button>
        )}
      </div>

      {activeTab === 'workflows' ? (
        <div className="workflows-content">
          <div className="workflows-main">
            <WorkflowList
              onExecute={handleExecuteWorkflow}
              onSelect={handleAddToBatch}
            />
          </div>
          {showBatchPanel && (
            <div className="batch-panel">
              <BatchExecution
                onComplete={() => {
                  setBatchItems([])
                  setShowBatchPanel(false)
                }}
              />
            </div>
          )}
        </div>
      ) : (
        <div className="templates-section">
          {loading ? (
            <div className="loading">加载中...</div>
          ) : (
            <div className="templates-grid">
              {templates.map((template: any) => (
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