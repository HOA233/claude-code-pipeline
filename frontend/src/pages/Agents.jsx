import React from 'react'
import { AgentList } from '../components/AgentList'
import './Agents.css'

function Agents() {
  return (
    <div className="agents-page">
      <div className="page-header">
        <h1>Agent 管理</h1>
        <p>管理 Claude Code Agent 配置</p>
      </div>
      <AgentList />
    </div>
  )
}

export default Agents