import React from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Layout from './components/Layout/Layout'
import ErrorBoundary from './components/ErrorBoundary/ErrorBoundary'
import NotificationProvider from './components/Notification/NotificationProvider'
import Dashboard from './pages/Dashboard'
import Skills from './pages/Skills'
import Tasks from './pages/Tasks'
import Pipelines from './pages/Pipelines'
import TaskPage from './pages/TaskPage'
import Settings from './pages/Settings'
import Agents from './pages/Agents'
import Workflows from './pages/Workflows'
import Executions from './pages/Executions'
import Schedules from './pages/Schedules'

function App() {
  return (
    <ErrorBoundary>
      <NotificationProvider>
        <BrowserRouter>
          <Layout>
            <Routes>
              <Route path="/" element={<Dashboard />} />
              <Route path="/skills" element={<Skills />} />
              <Route path="/tasks" element={<Tasks />} />
              <Route path="/tasks/new" element={<TaskPage />} />
              <Route path="/pipelines" element={<Pipelines />} />
              <Route path="/settings" element={<Settings />} />
              <Route path="/agents" element={<Agents />} />
              <Route path="/workflows" element={<Workflows />} />
              <Route path="/executions" element={<Executions />} />
              <Route path="/schedules" element={<Schedules />} />
            </Routes>
          </Layout>
        </BrowserRouter>
      </NotificationProvider>
    </ErrorBoundary>
  )
}

export default App