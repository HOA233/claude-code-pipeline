import React from 'react';
import { BrowserRouter as Router, Routes, Route, NavLink } from 'react-router-dom';
import { Dashboard } from './components/Dashboard';
import { AgentList } from './components/AgentList';
import { WorkflowList } from './components/WorkflowList';
import { ExecutionList } from './components/ExecutionList';
import { ScheduledJobList } from './components/ScheduledJobList';
import './App.css';

const App: React.FC = () => {
  return (
    <Router>
      <div className="app">
        <nav className="sidebar">
          <div className="logo">
            <h1>🤖 Agent Platform</h1>
          </div>
          <ul className="nav-menu">
            <li>
              <NavLink to="/" className={({ isActive }) => isActive ? 'active' : ''}>
                📊 仪表盘
              </NavLink>
            </li>
            <li>
              <NavLink to="/agents" className={({ isActive }) => isActive ? 'active' : ''}>
                🤖 Agent 管理
              </NavLink>
            </li>
            <li>
              <NavLink to="/workflows" className={({ isActive }) => isActive ? 'active' : ''}>
                🔄 工作流管理
              </NavLink>
            </li>
            <li>
              <NavLink to="/executions" className={({ isActive }) => isActive ? 'active' : ''}>
                📋 任务执行
              </NavLink>
            </li>
            <li>
              <NavLink to="/schedules" className={({ isActive }) => isActive ? 'active' : ''}>
                ⏰ 定时任务
              </NavLink>
            </li>
          </ul>
        </nav>

        <main className="main-content">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/agents" element={<AgentList />} />
            <Route path="/workflows" element={<WorkflowList />} />
            <Route path="/executions" element={<ExecutionList />} />
            <Route path="/schedules" element={<ScheduledJobList />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
};

export default App;