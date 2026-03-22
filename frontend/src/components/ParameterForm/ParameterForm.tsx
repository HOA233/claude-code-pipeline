import React from 'react'
import './ParameterForm.css'

const ParameterForm = ({ parameters = [], values = {}, onChange }) => {
  const handleChange = (paramName, value) => {
    onChange({
      ...values,
      [paramName]: value,
    })
  }

  const renderField = (param) => {
    const { name, type, required, description, default: defaultValue, values: enumValues } = param

    switch (type) {
      case 'string':
        return (
          <div key={name} className="form-field">
            <label className="field-label">
              {name}
              {required && <span className="required">*</span>}
            </label>
            {description && <p className="field-description">{description}</p>}
            <input
              type="text"
              className="field-input"
              value={values[name] ?? defaultValue ?? ''}
              onChange={(e) => handleChange(name, e.target.value)}
              placeholder={description || name}
            />
          </div>
        )

      case 'number':
      case 'integer':
        return (
          <div key={name} className="form-field">
            <label className="field-label">
              {name}
              {required && <span className="required">*</span>}
            </label>
            {description && <p className="field-description">{description}</p>}
            <input
              type="number"
              className="field-input"
              value={values[name] ?? defaultValue ?? ''}
              onChange={(e) => handleChange(name, e.target.value ? Number(e.target.value) : '')}
              placeholder={description || name}
            />
          </div>
        )

      case 'boolean':
        return (
          <div key={name} className="form-field">
            <label className="field-label">
              <input
                type="checkbox"
                className="field-checkbox"
                checked={values[name] ?? defaultValue ?? false}
                onChange={(e) => handleChange(name, e.target.checked)}
              />
              <span>
                {name}
                {required && <span className="required">*</span>}
              </span>
            </label>
            {description && <p className="field-description">{description}</p>}
          </div>
        )

      case 'enum':
      case 'select':
        return (
          <div key={name} className="form-field">
            <label className="field-label">
              {name}
              {required && <span className="required">*</span>}
            </label>
            {description && <p className="field-description">{description}</p>}
            <select
              className="field-select"
              value={values[name] ?? defaultValue ?? ''}
              onChange={(e) => handleChange(name, e.target.value)}
            >
              <option value="">Select {name}...</option>
              {(enumValues || []).map((val) => (
                <option key={val} value={val}>
                  {val}
                </option>
              ))}
            </select>
          </div>
        )

      case 'array':
        return (
          <div key={name} className="form-field">
            <label className="field-label">
              {name}
              {required && <span className="required">*</span>}
            </label>
            {description && <p className="field-description">{description}</p>}
            <input
              type="text"
              className="field-input"
              value={Array.isArray(values[name]) ? values[name].join(', ') : ''}
              onChange={(e) => handleChange(name, e.target.value.split(',').map(s => s.trim()))}
              placeholder="Comma-separated values"
            />
          </div>
        )

      case 'text':
      case 'textarea':
        return (
          <div key={name} className="form-field">
            <label className="field-label">
              {name}
              {required && <span className="required">*</span>}
            </label>
            {description && <p className="field-description">{description}</p>}
            <textarea
              className="field-textarea"
              value={values[name] ?? defaultValue ?? ''}
              onChange={(e) => handleChange(name, e.target.value)}
              placeholder={description || name}
              rows={4}
            />
          </div>
        )

      default:
        return (
          <div key={name} className="form-field">
            <label className="field-label">
              {name}
              {required && <span className="required">*</span>}
            </label>
            {description && <p className="field-description">{description}</p>}
            <input
              type="text"
              className="field-input"
              value={values[name] ?? defaultValue ?? ''}
              onChange={(e) => handleChange(name, e.target.value)}
              placeholder={description || name}
            />
          </div>
        )
    }
  }

  if (parameters.length === 0) {
    return (
      <div className="parameter-form-empty">
        <p>No parameters required for this skill.</p>
      </div>
    )
  }

  return (
    <div className="parameter-form">
      {parameters.map(renderField)}
    </div>
  )
}

export default ParameterForm