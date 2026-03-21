import React from 'react'
import './SkillCard.css'

const SkillCard = ({ skill, onSelect }) => {
  const getCategoryIcon = (category) => {
    const icons = {
      quality: '🔍',
      devops: '🚀',
      testing: '🧪',
      development: '🛠️',
      documentation: '📄',
    }
    return icons[category] || '⚡'
  }

  return (
    <div className="skill-card" onClick={() => onSelect?.(skill)}>
      <div className="skill-card-header">
        <span className="skill-icon">{getCategoryIcon(skill.category)}</span>
        <h3 className="skill-name">{skill.name}</h3>
      </div>

      <p className="skill-description">{skill.description}</p>

      <div className="skill-meta">
        <span className="skill-version">v{skill.version}</span>
        <span className={`skill-status ${skill.enabled ? 'enabled' : 'disabled'}`}>
          {skill.enabled ? 'Available' : 'Disabled'}
        </span>
      </div>

      <div className="skill-tags">
        {skill.tags?.slice(0, 3).map((tag, index) => (
          <span key={index} className="skill-tag">{tag}</span>
        ))}
      </div>

      <button className="skill-select-btn">
        Select Skill
      </button>
    </div>
  )
}

export default SkillCard