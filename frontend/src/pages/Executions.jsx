import React from 'react'
import { ExecutionList } from '../components/ExecutionList'
import { ExecutionCharts } from '../components/ExecutionCharts'
import './Executions.css'

function Executions() {
  return (
    <div className="executions-page">
      <div className="page-header">
        <h1>任务执行</h1>
        <p>查看和管理所有任务执行状态</p>
      </div>

      <div className="executions-content">
        <div className="charts-section">
          <ExecutionCharts />
        </div>

        <div className="list-section">
          <ExecutionList autoRefresh={true} />
        </div>
      </div>
    </div>
  )
}

export default Executions