import React from 'react'
import { useUIStore } from '../../stores'
import './ThemeToggle.css'

const ThemeToggle = () => {
  const { theme, setTheme } = useUIStore()

  const toggleTheme = () => {
    setTheme(theme === 'dark' ? 'light' : 'dark')
    document.documentElement.setAttribute('data-theme', theme === 'dark' ? 'light' : 'dark')
  }

  return (
    <button
      className="theme-toggle"
      onClick={toggleTheme}
      title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
    >
      <span className="toggle-track">
        <span className={`toggle-thumb ${theme}`}>
          {theme === 'dark' ? '🌙' : '☀️'}
        </span>
      </span>
    </button>
  )
}

export default ThemeToggle