import React, { useEffect, useState } from 'react'
import { tasksApi } from '../../api/tasks'
import './TaskList.css'

const TaskList = ({ onSelectTask }) => {
  const [tasks, setTasks] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadTasks()
  }, [])

  const loadTasks = async () => {
    try {
      const data = await tasksApi.getList()
      setTasks(data.tasks || [])
    } catch (err) {
      console.error('Failed to load tasks:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleCancel = async (taskId) => {
    try {
      await tasksApi.cancel(taskId)
      setTasks((prev) =>
        prev.map((t) => (t.id === taskId ? { ...t, status: 'cancelled' } : t))
      )
    } catch (err) {
      console.error('Failed to cancel task:', err)
    }
  }

  const getStatusIcon = (status) => {
    const icons = {
      pending: '⏳',
      running: '▶️',
      completed: '✅',
      failed: '❌',
      cancelled: '🚫',
    }
    return icons[status] || '❓'
  }

  if (loading) {
    return <div className="loading">Loading tasks...</div>
  }

  return (
    <div className="task-list">
      <div className="task-list-header">
        <h2>Tasks</h2>
        <span className="task-count">{tasks.length} tasks</span>
      </div>

      <div className="task-list-content">
        {tasks.length === 0 ? (
          <div className="empty-state">
            <div className="empty-state-icon">📋</div>
            <p className="empty-state-text">No tasks yet</p>
          </div>
        ) : (
          tasks.map((task) => (
            <div key={task.id} className="task-item">
              <div className={`task-status-indicator ${task.status}`}>
                {getStatusIcon(task.status)}
              </div>

              <div className="task-info">
                <div className="task-name">{task.id}</div>
                <div className="task-skill">{task.skill_id}</div>
              </div>

              <div className="task-status-badge">
                <span className={`status-badge ${task.status}`}>
                  {task.status}
                </span>
              </div>

              <div className="task-actions">
                <button
                  className="btn btn-ghost"
                  onClick={() => onSelectTask?.(task)}
                >
                  View
                </button>
                {task.status === 'running' && (
                  <button
                    className="btn btn-ghost"
                    onClick={() => handleCancel(task.id)}
                  >
                    Cancel
                  </button>
                )}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}

export default TaskList