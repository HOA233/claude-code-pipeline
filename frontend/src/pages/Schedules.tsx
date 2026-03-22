import React from 'react'
import { ScheduledJobList } from '../components/ScheduledJobList'
import './Schedules.css'

function Schedules() {
  return (
    <div className="schedules-page">
      <div className="page-header">
        <h1>定时任务</h1>
        <p>管理定时执行的 Agent 和工作流</p>
      </div>
      <ScheduledJobList />
    </div>
  )
}

export default Schedules