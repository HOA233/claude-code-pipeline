import React from 'react'
import './MessageBubble.css'

const MessageBubble = ({ type = 'assistant', content, timestamp }) => {
  const formatTime = (ts) => {
    if (!ts) return ''
    const date = new Date(ts)
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }

  return (
    <div className={`message-bubble ${type}`}>
      <div className="message-avatar">
        {type === 'assistant' ? (
          <span className="avatar-icon">✦</span>
        ) : (
          <span className="avatar-icon user">👤</span>
        )}
      </div>
      <div className="message-content">
        <div className="message-header">
          <span className="message-sender">
            {type === 'assistant' ? 'Claude' : 'You'}
          </span>
          {timestamp && (
            <span className="message-time">{formatTime(timestamp)}</span>
          )}
        </div>
        <div className="message-body">
          {typeof content === 'string' ? (
            <p>{content}</p>
          ) : (
            content
          )}
        </div>
      </div>
    </div>
  )
}

export default MessageBubble