import React, { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { skillsApi } from '../api/skills'
import { tasksApi } from '../api/tasks'
import { useTaskStore } from '../stores'
import ParameterForm from '../components/ParameterForm/ParameterForm'
import OutputConsole from '../components/OutputConsole/OutputConsole'
import './TaskPage.css'

const TaskPage = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const skillId = searchParams.get('skill')

  const { tasks, activeTask, setActiveTask, updateTask, appendOutput } = useTaskStore()
  const [skill, setSkill] = useState(null)
  const [loading, setLoading] = useState(false)
  const [params, setParams] = useState({})
  const [taskResult, setTaskResult] = useState(null)

  useEffect(() => {
    if (skillId) {
      loadSkill(skillId)
    }
  }, [skillId])

  const loadSkill = async (id) => {
    try {
      const data = await skillsApi.getDetail(id)
      setSkill(data)
      // Set default params
      const defaults = {}
      data.parameters?.forEach(p => {
        if (p.default !== undefined) {
          defaults[p.name] = p.default
        }
      })
      setParams(defaults)
    } catch (err) {
      console.error('Failed to load skill:', err)
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (!skill) return

    setLoading(true)
    try {
      const task = await tasksApi.create({
        skill_id: skill.id,
        parameters: params,
      })
      setActiveTask(task)
      navigate(`/tasks/${task.id}`)
    } catch (err) {
      console.error('Failed to create task:', err)
    } finally {
      setLoading(false)
    }
  }

  if (skillId && !skill) {
    return (
      <div className="task-page loading">
        <div className="loading-spinner"></div>
        <p>Loading skill configuration...</p>
      </div>
    )
  }

  return (
    <div className="task-page">
      <div className="task-header">
        <h1>Create Task</h1>
        {skill && (
          <div className="skill-badge">
            <span className="skill-icon">⚡</span>
            <span>{skill.name}</span>
          </div>
        )}
      </div>

      <div className="task-content">
        {skill ? (
          <form onSubmit={handleSubmit} className="task-form">
            <div className="form-section">
              <h2>Parameters</h2>
              <p className="section-description">
                Configure the parameters for {skill.name}
              </p>
              <ParameterForm
                parameters={skill.parameters || []}
                values={params}
                onChange={setParams}
              />
            </div>

            <div className="form-actions">
              <button
                type="button"
                className="btn-secondary"
                onClick={() => navigate('/')}
              >
                Cancel
              </button>
              <button
                type="submit"
                className="btn-primary"
                disabled={loading}
              >
                {loading ? 'Creating...' : 'Create Task'}
              </button>
            </div>
          </form>
        ) : (
          <div className="skill-selector">
            <h2>Select a Skill</h2>
            <p>Choose a skill to create a task</p>
            <div className="skill-list">
              <button onClick={() => navigate('/?selectSkill=true')}>
                Browse Skills
              </button>
            </div>
          </div>
        )}
      </div>

      {activeTask && (
        <div className="task-output">
          <h2>Task Output</h2>
          <OutputConsole
            taskId={activeTask.id}
            output={[]}
            status={activeTask.status}
          />
        </div>
      )}
    </div>
  )
}

export default TaskPage