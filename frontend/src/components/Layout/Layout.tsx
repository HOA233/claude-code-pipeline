import React from 'react'
import { Link, useLocation } from 'react-router-dom'
import './Layout.css'

const Layout = ({ children }) => {
  const location = useLocation()

  const navItems = [
    { path: '/', label: 'Dashboard', icon: '🏠' },
    { path: '/agents', label: 'Agents', icon: '🤖' },
    { path: '/workflows', label: 'Workflows', icon: '🔄' },
    { path: '/executions', label: 'Executions', icon: '📋' },
    { path: '/schedules', label: 'Schedules', icon: '⏰' },
    { path: '/skills', label: 'Skills', icon: '⚡' },
    { path: '/tasks', label: 'Tasks', icon: '📝' },
    { path: '/pipelines', label: 'Pipelines', icon: '🔗' },
    { path: '/settings', label: 'Settings', icon: '⚙️' },
  ]

  return (
    <div className="app-container">
      <aside className="sidebar">
        <div className="sidebar-header">
          <div className="sidebar-logo">
            <svg width="32" height="32" viewBox="0 0 40 40" fill="none">
              <defs>
                <linearGradient id="logoGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                  <stop offset="0%" stopColor="#D97706" />
                  <stop offset="100%" stopColor="#FB923C" />
                </linearGradient>
              </defs>
              <rect width="40" height="40" rx="10" fill="url(#logoGradient)" />
              <path
                d="M12 20C12 15.5817 15.5817 12 20 12V12C24.4183 12 28 15.5817 28 20V28H20C15.5817 28 12 24.4183 12 20V20Z"
                fill="white"
                fillOpacity="0.9"
              />
            </svg>
          </div>
          <h1 className="sidebar-title">Agent Platform</h1>
        </div>

        <nav className="sidebar-nav">
          {navItems.map((item) => (
            <Link
              key={item.path}
              to={item.path}
              className={`nav-item ${location.pathname === item.path ? 'active' : ''}`}
            >
              <span className="nav-icon">{item.icon}</span>
              <span className="nav-label">{item.label}</span>
            </Link>
          ))}
        </nav>

        <div className="sidebar-footer">
          <div className="status-indicator">
            <span className="status-dot"></span>
            <span className="status-text">System Online</span>
          </div>
        </div>
      </aside>

      <main className="main-content">
        {children}
      </main>
    </div>
  )
}

export default Layout