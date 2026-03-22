import React from 'react'
import { WorkflowList } from '../components/WorkflowList'
import './Workflows.css'

function Workflows() {
  return (
    <div className="workflows-page">
      <div className="page-header">
        <h1>工作流管理</h1>
        <p>配置和组合多个 Agent 形成工作流</p>
      </div>
      <WorkflowList />
    </div>
  )
}

export default Workflows