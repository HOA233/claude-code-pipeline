import React, { useEffect, useState } from 'react'
import { pipelinesApi } from '../api/pipelines'
import './Pipelines.css'

const Pipelines = () => {
  const [pipelines, setPipelines] = useState([])
  const [runs, setRuns] = useState([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [newPipeline, setNewPipeline] = useState({
    name: '',
    description: '',
    mode: 'serial',
    steps: [],
  })

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    try {
      const [pipelinesData, runsData] = await Promise.all([
        pipelinesApi.getList(),
        Promise.all(
          (pipelinesData?.pipelines || []).map(p =>
            pipelinesApi.getRuns(p.id).catch(() => ({ runs: [] }))
          )
        )
      ])
      setPipelines(pipelinesData.pipelines || [])
      setRuns(runsData.flat())
    } catch (err) {
      console.error('Failed to load data:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleRunPipeline = async (pipelineId) => {
    try {
      await pipelinesApi.run(pipelineId, {})
      loadData()
    } catch (err) {
      console.error('Failed to run pipeline:', err)
    }
  }

  const handleDeletePipeline = async (pipelineId) => {
    if (!confirm('Are you sure you want to delete this pipeline?')) return
    try {
      await pipelinesApi.delete(pipelineId)
      loadData()
    } catch (err) {
      console.error('Failed to delete pipeline:', err)
    }
  }

  const getModeIcon = (mode) => {
    const icons = {
      serial: '➡️',
      parallel: '⚡',
      hybrid: '🔀',
    }
    return icons[mode] || '❓'
  }

  if (loading) {
    return (
      <div className="pipelines-page loading">
        <div className="loading-spinner"></div>
        <p>Loading pipelines...</p>
      </div>
    )
  }

  return (
    <div className="pipelines-page">
      <div className="page-header">
        <div>
          <h1>Pipelines</h1>
          <p className="page-subtitle">
            Create and manage multi-CLI orchestration pipelines
          </p>
        </div>
        <button
          className="btn btn-primary"
          onClick={() => setShowCreate(true)}
        >
          Create Pipeline
        </button>
      </div>

      <div className="pipelines-stats">
        <div className="stat-card">
          <div className="stat-value">{pipelines.length}</div>
          <div className="stat-label">Pipelines</div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{runs.length}</div>
          <div className="stat-label">Total Runs</div>
        </div>
      </div>

      <div className="pipelines-list">
        {pipelines.length === 0 ? (
          <div className="empty-state">
            <div className="empty-state-icon">🔗</div>
            <p className="empty-state-text">No pipelines yet</p>
            <button className="btn btn-primary" onClick={() => setShowCreate(true)}>
              Create your first pipeline
            </button>
          </div>
        ) : (
          pipelines.map((pipeline) => (
            <div key={pipeline.id} className="pipeline-card">
              <div className="pipeline-header">
                <div className="pipeline-icon">{getModeIcon(pipeline.mode)}</div>
                <div className="pipeline-info">
                  <h3 className="pipeline-name">{pipeline.name}</h3>
                  <p className="pipeline-description">{pipeline.description}</p>
                </div>
                <div className="pipeline-mode">
                  <span className={`mode-badge ${pipeline.mode}`}>
                    {pipeline.mode}
                  </span>
                </div>
              </div>

              <div className="pipeline-steps">
                <h4>Steps ({pipeline.steps?.length || 0})</h4>
                <div className="steps-list">
                  {pipeline.steps?.map((step, index) => (
                    <div key={step.id || index} className="step-item">
                      <span className="step-number">{index + 1}</span>
                      <span className="step-name">{step.name || step.id}</span>
                      <span className="step-cli">{step.cli}</span>
                    </div>
                  ))}
                </div>
              </div>

              <div className="pipeline-actions">
                <button
                  className="btn btn-primary"
                  onClick={() => handleRunPipeline(pipeline.id)}
                >
                  Run Pipeline
                </button>
                <button
                  className="btn btn-secondary"
                  onClick={() => handleDeletePipeline(pipeline.id)}
                >
                  Delete
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}

export default Pipelines