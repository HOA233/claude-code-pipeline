import React from 'react'
import { useStatsStore } from '../../stores'
import './Stats.css'

const Stats = () => {
  const { stats } = useStatsStore()

  const statItems = [
    {
      label: 'Task Queue',
      value: stats.taskQueueLength,
      icon: '📋',
      color: '#D97706',
    },
    {
      label: 'Run Queue',
      value: stats.runQueueLength,
      icon: '🔄',
      color: '#D97706',
    },
    {
      label: 'Active Tasks',
      value: stats.activeTasks,
      icon: '⚡',
      color: '#10B981',
    },
    {
      label: 'Active Runs',
      value: stats.activeRuns,
      icon: '🏃',
      color: '#10B981',
    },
    {
      label: 'Completed',
      value: stats.completedTasks,
      icon: '✅',
      color: '#6366F1',
    },
    {
      label: 'Failed',
      value: stats.failedTasks,
      icon: '❌',
      color: '#EF4444',
    },
  ]

  return (
    <div className="stats-container">
      {statItems.map((item, index) => (
        <div key={index} className="stat-card">
          <div className="stat-icon" style={{ backgroundColor: `${item.color}20` }}>
            <span>{item.icon}</span>
          </div>
          <div className="stat-content">
            <span className="stat-value" style={{ color: item.color }}>
              {item.value}
            </span>
            <span className="stat-label">{item.label}</span>
          </div>
        </div>
      ))}
    </div>
  )
}

export default Stats