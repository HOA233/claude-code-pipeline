import React from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Layout from './components/Layout/Layout'
import Dashboard from './pages/Dashboard'
import Skills from './pages/Skills'
import Tasks from './pages/Tasks'
import Pipelines from './pages/Pipelines'

function App() {
  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/skills" element={<Skills />} />
          <Route path="/tasks" element={<Tasks />} />
          <Route path="/pipelines" element={<Pipelines />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  )
}

export default App