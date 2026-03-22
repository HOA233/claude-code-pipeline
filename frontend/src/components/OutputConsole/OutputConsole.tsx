import React, { useEffect, useRef, useState } from 'react'
import './OutputConsole.css'

const OutputConsole = ({ taskId, output = [], status = 'idle' }) => {
  const consoleRef = useRef(null)
  const [autoScroll, setAutoScroll] = useState(true)

  // Auto-scroll to bottom when new output arrives
  useEffect(() => {
    if (autoScroll && consoleRef.current) {
      consoleRef.current.scrollTop = consoleRef.current.scrollHeight
    }
  }, [output, autoScroll])

  const handleScroll = () => {
    if (consoleRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = consoleRef.current
      const isAtBottom = scrollHeight - scrollTop - clientHeight < 50
      setAutoScroll(isAtBottom)
    }
  }

  const getStatusIndicator = () => {
    switch (status) {
      case 'running':
        return <span className="status-indicator running">● Running</span>
      case 'completed':
        return <span className="status-indicator completed">✓ Completed</span>
      case 'failed':
        return <span className="status-indicator failed">✗ Failed</span>
      case 'pending':
        return <span className="status-indicator pending">○ Pending</span>
      default:
        return <span className="status-indicator idle">○ Idle</span>
    }
  }

  return (
    <div className="output-console">
      <div className="console-header">
        <div className="console-title">
          <span className="console-icon">{'>'}</span>
          <span>Console Output</span>
          {getStatusIndicator()}
        </div>
        <div className="console-actions">
          <button
            className={`action-btn ${autoScroll ? 'active' : ''}`}
            onClick={() => setAutoScroll(!autoScroll)}
            title="Auto-scroll"
          >
            ⤓
          </button>
          <button
            className="action-btn"
            onClick={() => {
              if (consoleRef.current) {
                consoleRef.current.scrollTop = 0
              }
            }}
            title="Scroll to top"
          >
            ⤒
          </button>
          <button
            className="action-btn"
            onClick={() => navigator.clipboard?.writeText(output.join('\n'))}
            title="Copy all"
          >
            📋
          </button>
        </div>
      </div>

      <div
        ref={consoleRef}
        className="console-content"
        onScroll={handleScroll}
      >
        {output.length === 0 ? (
          <div className="console-placeholder">
            {status === 'pending' && 'Waiting for task to start...'}
            {status === 'running' && 'Starting output...'}
            {status === 'idle' && 'No output available'}
            {status === 'completed' && 'Task completed - no output'}
            {status === 'failed' && 'Task failed - no output'}
          </div>
        ) : (
          output.map((line, index) => (
            <div key={index} className="console-line">
              <span className="line-number">{index + 1}</span>
              <span className="line-content">{line}</span>
            </div>
          ))
        )}
        {status === 'running' && output.length > 0 && (
          <div className="console-cursor">
            <span className="cursor">▋</span>
          </div>
        )}
      </div>
    </div>
  )
}

export default OutputConsole