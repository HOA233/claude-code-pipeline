import React from 'react'
import './Loading.css'

export const LoadingSpinner = ({ size = 24, className = '' }) => (
  <div className={`loading-spinner ${className}`} style={{ width: size, height: size }} />
)

export const LoadingDots = ({ className = '' }) => (
  <div className={`loading-dots ${className}`}>
    <span></span>
    <span></span>
    <span></span>
  </div>
)

export const LoadingSkeleton = ({ width = '100%', height = 20, rounded = false, className = '' }) => (
  <div
    className={`loading-skeleton ${rounded ? 'rounded' : ''} ${className}`}
    style={{ width, height }}
  />
)

export const LoadingCard = ({ className = '' }) => (
  <div className={`loading-card ${className}`}>
    <LoadingSkeleton height={16} width="60%" />
    <LoadingSkeleton height={12} width="80%" />
    <LoadingSkeleton height={12} width="40%" />
  </div>
)

export const LoadingList = ({ count = 5, className = '' }) => (
  <div className={`loading-list ${className}`}>
    {Array.from({ length: count }).map((_, i) => (
      <LoadingCard key={i} />
    ))}
  </div>
)

export const LoadingPage = ({ message = 'Loading...', className = '' }) => (
  <div className={`loading-page ${className}`}>
    <LoadingSpinner size={40} />
    <p className="loading-message">{message}</p>
  </div>
)

export const LoadingOverlay = ({ visible, message }) => {
  if (!visible) return null
  return (
    <div className="loading-overlay">
      <div className="loading-overlay-content">
        <LoadingSpinner size={32} />
        {message && <p>{message}</p>}
      </div>
    </div>
  )
}

const Loading = ({ type = 'spinner', ...props }) => {
  switch (type) {
    case 'spinner':
      return <LoadingSpinner {...props} />
    case 'dots':
      return <LoadingDots {...props} />
    case 'skeleton':
      return <LoadingSkeleton {...props} />
    case 'card':
      return <LoadingCard {...props} />
    case 'list':
      return <LoadingList {...props} />
    case 'page':
      return <LoadingPage {...props} />
    case 'overlay':
      return <LoadingOverlay {...props} />
    default:
      return <LoadingSpinner {...props} />
  }
}

export default Loading