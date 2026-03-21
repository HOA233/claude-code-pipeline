import React, { useState } from 'react'
import { tasksApi } from '../../api/tasks'
import './CreateTask.css'

const CreateTask = ({ skill, onClose, onSubmit }) => {
  const [parameters, setParameters] = useState({})
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  // Initialize default parameter values
  React.useEffect(() => {
    const defaults = {}
    skill.parameters?.forEach((param) => {
      if (param.default !== undefined) {
        defaults[param.name] = param.default
      }
    })
    setParameters(defaults)
  }, [skill])

  const handleSubmit = async () => {
    setLoading(true)
    setError(null)

    try {
      const task = await tasksApi.create({
        skill_id: skill.id,
        parameters,
        options: {
          timeout: skill.cli?.timeout || 600,
        },
      })

      onSubmit?.(task)
      onClose()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to create task')
    } finally {
      setLoading(false)
    }
  }

  const renderField = (param) => {
    const { name, type, required, description, values } = param

    switch (type) {
      case 'string':
        return (
          <input
            type="text"
            className="form-input"
            value={parameters[name] || ''}
            onChange={(e) => setParameters({ ...parameters, [name]: e.target.value })}
            placeholder={description}
            required={required}
          />
        )

      case 'enum':
        return (
          <select
            className="form-select"
            value={parameters[name] || ''}
            onChange={(e) => setParameters({ ...parameters, [name]: e.target.value })}
            required={required}
          >
            <option value="">Select...</option>
            {values?.map((v) => (
              <option key={v} value={v}>{v}</option>
            ))}
          </select>
        )

      case 'boolean':
        return (
          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={parameters[name] || false}
              onChange={(e) => setParameters({ ...parameters, [name]: e.target.checked })}
            />
            <span className="checkbox-text">{description}</span>
          </label>
        )

      case 'number':
        return (
          <input
            type="number"
            className="form-input"
            value={parameters[name] || ''}
            onChange={(e) => setParameters({ ...parameters, [name]: Number(e.target.value) })}
            placeholder={description}
            required={required}
          />
        )

      default:
        return null
    }
  }

  return (
    <div className="modal-overlay">
      <div className="modal-content">
        <div className="modal-header">
          <h2 className="modal-title">Create Task: {skill.name}</h2>
          <button className="modal-close" onClick={onClose}>×</button>
        </div>

        <div className="modal-body">
          <div className="skill-info">
            <p>{skill.description}</p>
          </div>

          <div className="parameter-form">
            {skill.parameters?.map((param) => (
              <div key={param.name} className="form-group">
                <label className="form-label">
                  {param.name}
                  {param.required && <span className="required">*</span>}
                </label>
                {renderField(param)}
                {param.description && (
                  <small className="form-hint">{param.description}</small>
                )}
              </div>
            ))}
          </div>

          {error && (
            <div className="error-message">{error}</div>
          )}
        </div>

        <div className="modal-footer">
          <button className="btn btn-secondary" onClick={onClose}>
            Cancel
          </button>
          <button
            className="btn btn-primary"
            onClick={handleSubmit}
            disabled={loading}
          >
            {loading ? 'Creating...' : 'Create Task'}
          </button>
        </div>
      </div>
    </div>
  )
}

export default CreateTask