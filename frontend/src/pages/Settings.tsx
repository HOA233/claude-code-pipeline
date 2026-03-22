import React, { useState } from 'react'
import { useUIStore } from '../stores'
import ThemeToggle from '../components/ThemeToggle/ThemeToggle'
import './Settings.css'

const Settings = () => {
  const { theme, setTheme } = useUIStore()
  const [apiUrl, setApiUrl] = useState(localStorage.getItem('apiUrl') || 'http://localhost:8080')
  const [saved, setSaved] = useState(false)

  const handleSave = () => {
    localStorage.setItem('apiUrl', apiUrl)
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  const handleClearCache = () => {
    if (confirm('Are you sure you want to clear all local data?')) {
      localStorage.clear()
      window.location.reload()
    }
  }

  const handleExportSettings = () => {
    const settings = {
      theme,
      apiUrl,
      exportedAt: new Date().toISOString(),
    }
    const blob = new Blob([JSON.stringify(settings, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'claude-pipeline-settings.json'
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="settings-page">
      <div className="settings-header">
        <h1>Settings</h1>
        <p className="settings-subtitle">Configure your Claude Pipeline preferences</p>
      </div>

      <div className="settings-content">
        <section className="settings-section">
          <h2>Appearance</h2>
          <div className="settings-item">
            <div className="item-info">
              <span className="item-label">Theme</span>
              <span className="item-description">Choose between light and dark mode</span>
            </div>
            <ThemeToggle />
          </div>
        </section>

        <section className="settings-section">
          <h2>Connection</h2>
          <div className="settings-item">
            <div className="item-info">
              <span className="item-label">API URL</span>
              <span className="item-description">The base URL for the Claude Pipeline API</span>
            </div>
            <input
              type="text"
              className="settings-input"
              value={apiUrl}
              onChange={(e) => setApiUrl(e.target.value)}
              placeholder="http://localhost:8080"
            />
          </div>
          <div className="settings-actions">
            <button className="btn-primary" onClick={handleSave}>
              {saved ? '✓ Saved' : 'Save Changes'}
            </button>
          </div>
        </section>

        <section className="settings-section">
          <h2>Data</h2>
          <div className="settings-item">
            <div className="item-info">
              <span className="item-label">Export Settings</span>
              <span className="item-description">Download your settings as a JSON file</span>
            </div>
            <button className="btn-secondary" onClick={handleExportSettings}>
              Export
            </button>
          </div>
          <div className="settings-item danger">
            <div className="item-info">
              <span className="item-label">Clear Local Data</span>
              <span className="item-description">Remove all cached data and settings</span>
            </div>
            <button className="btn-danger" onClick={handleClearCache}>
              Clear All
            </button>
          </div>
        </section>

        <section className="settings-section">
          <h2>About</h2>
          <div className="about-info">
            <div className="about-item">
              <span className="about-label">Version</span>
              <span className="about-value">1.0.0</span>
            </div>
            <div className="about-item">
              <span className="about-label">Repository</span>
              <a href="https://github.com/HOA233/claude-code-pipeline" target="_blank" rel="noopener noreferrer">
                github.com/HOA233/claude-code-pipeline
              </a>
            </div>
            <div className="about-item">
              <span className="about-label">License</span>
              <span className="about-value">MIT</span>
            </div>
          </div>
        </section>
      </div>
    </div>
  )
}

export default Settings