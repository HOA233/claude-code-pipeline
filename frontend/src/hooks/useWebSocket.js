import { useEffect, useRef, useCallback } from 'react'

/**
 * WebSocket hook for real-time updates
 * @param {string} url - WebSocket URL (e.g., 'ws://localhost:8080/ws/tasks/123')
 * @param {object} options - Options for the WebSocket connection
 * @param {function} options.onMessage - Callback for incoming messages
 * @param {function} options.onOpen - Callback when connection opens
 * @param {function} options.onClose - Callback when connection closes
 * @param {function} options.onError - Callback for errors
 * @param {number} options.reconnectInterval - Reconnect interval in ms (default: 3000)
 * @param {number} options.maxReconnectAttempts - Max reconnect attempts (default: 5)
 */
export function useWebSocket(url, options = {}) {
  const {
    onMessage,
    onOpen,
    onClose,
    onError,
    reconnectInterval = 3000,
    maxReconnectAttempts = 5,
  } = options

  const wsRef = useRef(null)
  const reconnectAttemptsRef = useRef(0)
  const reconnectTimeoutRef = useRef(null)
  const mountedRef = useRef(true)

  const connect = useCallback(() => {
    if (!url) return

    const ws = new WebSocket(url)
    wsRef.current = ws

    ws.onopen = () => {
      console.log('WebSocket connected:', url)
      reconnectAttemptsRef.current = 0
      if (mountedRef.current && onOpen) {
        onOpen()
      }
    }

    ws.onmessage = (event) => {
      if (mountedRef.current && onMessage) {
        try {
          const data = JSON.parse(event.data)
          onMessage(data)
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err)
          onMessage(event.data)
        }
      }
    }

    ws.onclose = () => {
      console.log('WebSocket closed:', url)
      if (mountedRef.current && onClose) {
        onClose()
      }

      // Attempt to reconnect
      if (
        mountedRef.current &&
        reconnectAttemptsRef.current < maxReconnectAttempts
      ) {
        reconnectAttemptsRef.current++
        console.log(
          `Reconnecting in ${reconnectInterval}ms (attempt ${reconnectAttemptsRef.current}/${maxReconnectAttempts})`
        )
        reconnectTimeoutRef.current = setTimeout(connect, reconnectInterval)
      }
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
      if (mountedRef.current && onError) {
        onError(error)
      }
    }
  }, [url, onMessage, onOpen, onClose, onError, reconnectInterval, maxReconnectAttempts])

  const disconnect = useCallback(() => {
    mountedRef.current = false
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
    }
    if (wsRef.current) {
      wsRef.current.close()
    }
  }, [])

  const send = useCallback((data) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      const message = typeof data === 'string' ? data : JSON.stringify(data)
      wsRef.current.send(message)
    }
  }, [])

  useEffect(() => {
    mountedRef.current = true
    connect()

    return () => {
      disconnect()
    }
  }, [connect, disconnect])

  return {
    send,
    disconnect,
    reconnect: connect,
    isConnected: wsRef.current?.readyState === WebSocket.OPEN,
  }
}

/**
 * Hook for subscribing to task updates
 * @param {string} taskId - Task ID to subscribe to
 * @param {function} onUpdate - Callback for task updates
 */
export function useTaskUpdates(taskId, onUpdate) {
  const baseUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:8080'
  const url = taskId ? `${baseUrl}/ws/tasks/${taskId}` : null

  return useWebSocket(url, {
    onMessage: onUpdate,
  })
}

/**
 * Hook for subscribing to run updates
 * @param {string} runId - Run ID to subscribe to
 * @param {function} onUpdate - Callback for run updates
 */
export function useRunUpdates(runId, onUpdate) {
  const baseUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:8080'
  const url = runId ? `${baseUrl}/ws/runs/${runId}` : null

  return useWebSocket(url, {
    onMessage: onUpdate,
  })
}

/**
 * Hook for subscribing to global system updates
 * @param {function} onUpdate - Callback for system updates
 */
export function useSystemUpdates(onUpdate) {
  const baseUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:8080'
  const url = `${baseUrl}/ws`

  return useWebSocket(url, {
    onMessage: onUpdate,
  })
}

export default useWebSocket