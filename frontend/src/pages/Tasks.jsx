import React from 'react'
import TaskList from '../components/TaskList/TaskList'
import './Tasks.css'

const Tasks = () => {
  return (
    <div className="tasks-page">
      <div className="page-header">
        <div>
          <h1>Tasks</h1>
          <p className="page-subtitle">
            View and manage your task execution history
          </p>
        </div>
      </div>

      <TaskList onSelectTask={(task) => console.log('View task:', task)} />
    </div>
  )
}

export default Tasks