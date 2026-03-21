import React, { useEffect, useState } from 'react'
import { skillsApi } from '../api/skills'
import SkillCard from '../components/SkillCard/SkillCard'
import './Skills.css'

const Skills = () => {
  const [skills, setSkills] = useState([])
  const [loading, setLoading] = useState(true)
  const [syncing, setSyncing] = useState(false)

  useEffect(() => {
    loadSkills()
  }, [])

  const loadSkills = async () => {
    try {
      const data = await skillsApi.getList()
      setSkills(data.skills || [])
    } catch (err) {
      console.error('Failed to load skills:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleSync = async () => {
    setSyncing(true)
    try {
      const data = await skillsApi.sync()
      console.log('Synced skills:', data)
      await loadSkills()
    } catch (err) {
      console.error('Failed to sync skills:', err)
    } finally {
      setSyncing(false)
    }
  }

  if (loading) {
    return (
      <div className="skills-page loading">
        <div className="loading-spinner"></div>
        <p>Loading skills...</p>
      </div>
    )
  }

  return (
    <div className="skills-page">
      <div className="page-header">
        <div>
          <h1>Skills</h1>
          <p className="page-subtitle">
            Manage and execute AI-powered skills
          </p>
        </div>
        <button
          className="btn btn-primary"
          onClick={handleSync}
          disabled={syncing}
        >
          {syncing ? 'Syncing...' : 'Sync from GitLab'}
        </button>
      </div>

      <div className="skills-stats">
        <div className="stat-card">
          <div className="stat-value">{skills.length}</div>
          <div className="stat-label">Total Skills</div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{skills.filter(s => s.enabled).length}</div>
          <div className="stat-label">Enabled</div>
        </div>
        <div className="stat-card">
          <div className="stat-value">
            {new Set(skills.map(s => s.category)).size}
          </div>
          <div className="stat-label">Categories</div>
        </div>
      </div>

      <div className="skills-grid">
        {skills.map((skill) => (
          <SkillCard key={skill.id} skill={skill} />
        ))}
      </div>
    </div>
  )
}

export default Skills