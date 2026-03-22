import React, { createContext, useContext, useCallback } from 'react'
import { useUIStore } from '../../stores'
import './NotificationProvider.css'

const NotificationContext = createContext(null)

export const useNotification = () => {
  const context = useContext(NotificationContext)
  if (!context) {
    throw new Error('useNotification must be used within NotificationProvider')
  }
  return context
}

const NotificationProvider = ({ children }) => {
  const { notifications, addNotification, removeNotification } = useUIStore()

  const show = useCallback((notification) => {
    const id = Date.now() + Math.random()
    const finalNotification = {
      id,
      type: 'info',
      duration: 5000,
      ...notification,
    }

    addNotification(finalNotification)

    if (finalNotification.duration > 0) {
      setTimeout(() => {
        removeNotification(id)
      }, finalNotification.duration)
    }

    return id
  }, [addNotification, removeNotification])

  const success = useCallback((message, options = {}) => {
    return show({ type: 'success', message, ...options })
  }, [show])

  const error = useCallback((message, options = {}) => {
    return show({ type: 'error', message, duration: 8000, ...options })
  }, [show])

  const warning = useCallback((message, options = {}) => {
    return show({ type: 'warning', message, ...options })
  }, [show])

  const info = useCallback((message, options = {}) => {
    return show({ type: 'info', message, ...options })
  }, [show])

  const dismiss = useCallback((id) => {
    removeNotification(id)
  }, [removeNotification])

  return (
    <NotificationContext.Provider value={{ show, success, error, warning, info, dismiss }}>
      {children}
      <div className="notification-container">
        {notifications.map((notification) => (
          <div
            key={notification.id}
            className={`notification notification-${notification.type}`}
            onClick={() => dismiss(notification.id)}
          >
            <div className="notification-icon">
              {notification.type === 'success' && '✓'}
              {notification.type === 'error' && '✗'}
              {notification.type === 'warning' && '⚠'}
              {notification.type === 'info' && 'ℹ'}
            </div>
            <div className="notification-content">
              {notification.title && (
                <div className="notification-title">{notification.title}</div>
              )}
              <div className="notification-message">{notification.message}</div>
            </div>
            <button
              className="notification-close"
              onClick={(e) => {
                e.stopPropagation()
                dismiss(notification.id)
              }}
            >
              ×
            </button>
          </div>
        ))}
      </div>
    </NotificationContext.Provider>
  )
}

export default NotificationProvider