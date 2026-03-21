import React, { useState, useEffect } from 'react'
import { pipelinesApi } from '../../api/pipelines'
import { useRunUpdates } from '../../hooks/useWebSocket'
import { usePipelineStore } from '../../stores'
import OutputConsole from '../OutputConsole/OutputConsole'
import './PipelineDetail.css'

const PipelineDetail = ({ pipelineId, onClose }) => {
  const { pipelines, runs, updatePipeline, addRun, updateRun } = usePipelineStore()
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState('overview')

  const pipeline = pipelines.find(p => p.id === pipelineId)
  const pipelineRuns = runs.filter(r => r.pipeline_id === pipelineId)

  useEffect(() => {
    loadPipeline()
  }, [pipelineId])

  const loadPipeline = async () => {
    try {
      const data = await pipelinesApi.getDetail(pipelineId)
      updatePipeline(pipelineId, data)
    } catch (err) {
      console.error('Failed to load pipeline:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleRun = async () => {
    try {
      const result = await pipelinesApi.run(pipelineId)
      addRun({
        id: result.run_id || result.id,
        pipeline_id: pipelineId,
        status: 'running',
        created_at: new Date().toISOString(),
      })
    } catch (err) {
      console.error('Failed to run pipeline:', err)
    }
  }

  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this pipeline?')) return
    try {
      await pipelinesApi.delete(pipelineId)
      onClose?.()
    } catch (err) {
      console.error('Failed to delete pipeline:', err)
    }
  }

  const getStatusColor = (status) => {
    switch (status) {
      case 'completed': return '#10B981'
      case 'running': return '#D97706'
      case 'failed': return '#EF4444'
      case 'pending': return '#FBBF24'
      default: return '#6B7280'
    }
  }

  if (loading) {
    return (
      <div className="pipeline-detail-overlay">
        <div className="pipeline-detail-modal loading">
          <div className="loading-spinner"></div>
        </div>
      </div>
    )
  }

  if (!pipeline) {
    return (
      <div className="pipeline-detail-overlay">
        <div className="pipeline-detail-modal">
          <p>Pipeline not found</p>
          <button onClick={onClose}>Close</button>
        </div>
      </div>
    )
  }

  return (
    <div className="pipeline-detail-overlay" onClick={onClose}>
      <div className="pipeline-detail-modal" onClick={e => e.stopPropagation()}>
        <div className="pipeline-detail-header">
          <div className="pipeline-info">
            <h2>{pipeline.name}</h2>
            <span className="pipeline-mode">{pipeline.mode}</span>
          </div>
          <button className="close-button" onClick={onClose}>×</button>
        </div>

        <div className="pipeline-tabs">
          <button
            className={`tab ${activeTab === 'overview' ? 'active' : ''}`}
            onClick={() => setActiveTab('overview')}
          >
            Overview
          </button>
          <button
            className={`tab ${activeTab === 'steps' ? 'active' : ''}`}
            onClick={() => setActiveTab('steps')}
          >
            Steps ({pipeline.steps?.length || 0})
          </button>
          <button
            className={`tab ${activeTab === 'runs' ? 'active' : ''}`}
            onClick={() => setActiveTab('runs')}
          >
            Runs ({pipelineRuns.length})
          </button>
        </div>

        <div className="pipeline-detail-content">
          {activeTab === 'overview' && (
            <div className="tab-content">
              {pipeline.description && (
                <p className="pipeline-description">{pipeline.description}</p>
              )}

              <div className="pipeline-meta">
                <div className="meta-item">
                  <span className="meta-label">Mode</span>
                  <span className="meta-value">{pipeline.mode}</span>
                </div>
                <div className="meta-item">
                  <span className="meta-label">Steps</span>
                  <span className="meta-value">{pipeline.steps?.length || 0}</span>
                </div>
                <div className="meta-item">
                  <span className="meta-label">Created</span>
                  <span className="meta-value">
                    {new Date(pipeline.created_at).toLocaleDateString()}
                  </span>
                </div>
              </div>

              <div className="pipeline-actions">
                <button className="btn-primary" onClick={handleRun}>
                  ▶ Run Pipeline
                </button>
                <button className="btn-secondary" onClick={handleDelete}>
                  🗑 Delete
                </button>
              </div>
            </div>
          )}

          {activeTab === 'steps' && (
            <div className="tab-content">
              <div className="steps-list">
                {(pipeline.steps || []).map((step, index) => (
                  <div key={step.id || index} className="step-item">
                    <div className="step-header">
                      <span className="step-number">{index + 1}</span>
                      <span className="step-name">{step.name || step.id}</span>
                      <span className="step-cli">{step.cli}</span>
                    </div>
                    <div className="step-details">
                      <span>Action: {step.action}</span>
                      {step.timeout && <span>Timeout: {step.timeout}s</span>}
                      {step.depends_on?.length > 0 && (
                        <span>Depends: {step.depends_on.join(', ')}</span>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {activeTab === 'runs' && (
            <div className="tab-content">
              {pipelineRuns.length === 0 ? (
                <div className="empty-state">
                  <p>No runs yet</p>
                  <button className="btn-primary" onClick={handleRun}>
                    Start First Run
                  </button>
                </div>
              ) : (
                <div className="runs-list">
                  {pipelineRuns.map(run => (
                    <div key={run.id} className="run-item">
                      <div className="run-status" style={{ backgroundColor: getStatusColor(run.status) }}>
                        {run.status}
                      </div>
                      <div className="run-info">
                        <span className="run-id">{run.id.slice(0, 8)}</span>
                        <span className="run-time">
                          {new Date(run.created_at).toLocaleString()}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default PipelineDetail