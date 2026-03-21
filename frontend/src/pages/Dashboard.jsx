import React, { useEffect, useState } from 'react'
import { skillsApi } from '../api/skills'
import SkillCard from '../components/SkillCard/SkillCard'
import TaskList from '../components/TaskList/TaskList'
import CreateTask from '../components/CreateTask/CreateTask'
import './Dashboard.css'

const Dashboard = () => {
  const [skills, setSkills] = useState([])
  const [loading, setLoading] = useState(true)
  const [selectedSkill, setSelectedSkill] = useState(null)
  const [showModal, setShowModal] = useState(false)

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

  const handleSkillSelect = (skill) => {
    setSelectedSkill(skill)
    setShowModal(true)
  }

  const handleTaskCreated = (task) => {
    console.log('Task created:', task)
  }

  if (loading) {
    return (
      <div className="dashboard-loading">
        <div className="loading-spinner"></div>
        <p>Loading dashboard...</p>
      </div>
    )
  }

  return (
    <div className="dashboard">
      <div className="dashboard-header">
        <h1>Dashboard</h1>
        <p className="dashboard-subtitle">
          Select a skill to create a task, or view your running tasks
        </p>
      </div>

      <section className="dashboard-section">
        <h2 className="section-title">Available Skills</h2>
        <div className="skills-grid">
          {skills.map((skill) => (
            <SkillCard
              key={skill.id}
              skill={skill}
              onSelect={handleSkillSelect}
            />
          ))}
        </div>
      </section>

      <section className="dashboard-section">
        <h2 className="section-title">Recent Tasks</h2>
        <TaskList onSelectTask={(task) => console.log('View task:', task)} />
      </section>

      {showModal && selectedSkill && (
        <CreateTask
          skill={selectedSkill}
          onClose={() => setShowModal(false)}
          onSubmit={handleTaskCreated}
        />
      )}
    </div>
  )
}

export default Dashboard