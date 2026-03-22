import React from 'react'
import { ExecutionList } from '../components/ExecutionList'
import './Executions.css'

function Executions() {
  return (
    <div className="executions-page">
      <div className="page-header">
        <h1>任务执行</h1>
        <p>查看和管理所有任务执行状态</p>
      </div>
      <ExecutionList autoRefresh={true} />
    </div>
  )
}

export default Executions