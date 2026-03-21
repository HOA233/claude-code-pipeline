import React, { useEffect, useRef, useState } from 'react'
import { useTaskUpdates } from '../../hooks/useWebSocket'
import { useTaskStore } from '../../stores'
import './TaskDetail.css'

const TaskDetail = ({ taskId, onClose }) => {
  const { tasks, taskOutput, updateTask, appendOutput } = useTaskStore()
  const task = tasks.find((t) => t.id === taskId) || {}
  const output = taskOutput[taskId] || []
  const outputRef = useRef(null)
  const [autoScroll, setAutoScroll] = useState(true)

  // Subscribe to task updates
  useTaskUpdates(taskId, (data) => {
    console.log('Task update:', data)

    switch (data.type) {
      case 'task:status':
        updateTask(taskId, data.data)
        break
      case 'task:output':
        appendOutput(taskId, data.data)
        break
      case 'task:progress':
        updateTask(taskId, { progress: data.progress })
        break
      case 'task:completed':
        updateTask(taskId, { status: 'completed', result: data.data })
        break
      case 'task:failed':
        updateTask(taskId, { status: 'failed', error: data.data?.error })
        break
      default:
        console.log('Unknown update type:', data.type)
    }
  })

  // Auto-scroll to bottom
  useEffect(() => {
    if (autoScroll && outputRef.current) {
      outputRef.current.scrollTop = outputRef.current.scrollHeight
    }
  }, [output, autoScroll])

  const handleScroll = () => {
    if (outputRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = outputRef.current
      const isAtBottom = scrollHeight - scrollTop - clientHeight < 50
      setAutoScroll(isAtBottom)
    }
  }

  const getStatusColor = (status) => {
    switch (status) {
      case 'pending':
        return '#FCD34D'
      case 'running':
        return '#D97706'
      case 'completed':
        return '#10B981'
      case 'failed':
        return '#EF4444'
      default:
        return '#6B7280'
    }
  }

  const formatDate = (date) => {
    if (!date) return 'N/A'
    return new Date(date).toLocaleString()
  }

  return (
    <div className="task-detail-overlay" onClick={onClose}>
      <div className="task-detail-modal" onClick={(e) => e.stopPropagation()}>
        <div className="task-detail-header">
          <div className="task-info">
            <h2>Task: {taskId.slice(0, 8)}...</h2>
            <div
              className="task-status-badge"
              style={{ backgroundColor: getStatusColor(task.status) }}
            >
              {task.status || 'unknown'}
            </div>
          </div>
          <button className="close-button" onClick={onClose}>
            ×
          </button>
        </div>

        <div className="task-meta">
          <div className="meta-item">
            <span className="meta-label">Skill</span>
            <span className="meta-value">{task.skill_id || 'N/A'}</span>
          </div>
          <div className="meta-item">
            <span className="meta-label">Created</span>
            <span className="meta-value">{formatDate(task.created_at)}</span>
          </div>
          <div className="meta-item">
            <span className="meta-label">Progress</span>
            <span className="meta-value">{task.progress || 0}%</span>
          </div>
        </div>

        {task.status === 'running' && (
          <div className="progress-bar">
            <div
              className="progress-fill"
              style={{ width: `${task.progress || 0}%` }}
            />
          </div>
        )}

        <div className="output-container">
          <div className="output-header">
            <span>Output</span>
            <button
              className="clear-output-btn"
              onClick={() => setAutoScroll(!autoScroll)}
            >
              {autoScroll ? 'Auto-scroll ON' : 'Auto-scroll OFF'}
            </button>
          </div>
          <div
            ref={outputRef}
            className="output-content"
            onScroll={handleScroll}
          >
            {output.length === 0 ? (
              <div className="output-placeholder">
                Waiting for output...
              </div>
            ) : (
              output.map((line, index) => (
                <div key={index} className="output-line">
                  {line}
                </div>
              ))
            )}
          </div>
        </div>

        {task.status === 'completed' && task.result && (
          <div className="result-container">
            <h3>Result</h3>
            <pre>{JSON.stringify(task.result, null, 2)}</pre>
          </div>
        )}

        {task.status === 'failed' && task.error && (
          <div className="error-container">
            <h3>Error</h3>
            <p>{task.error}</p>
          </div>
        )}
      </div>
    </div>
  )
}

export default TaskDetail