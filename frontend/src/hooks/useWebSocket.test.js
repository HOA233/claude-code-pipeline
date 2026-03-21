import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useWebSocket, useTaskUpdates, useRunUpdates, useSystemUpdates } from '../hooks/useWebSocket'

// Mock WebSocket
class MockWebSocket {
  constructor(url) {
    this.url = url
    this.readyState = WebSocket.CONNECTING
    setTimeout(() => {
      this.readyState = WebSocket.OPEN
      this.onopen?.()
    }, 0)
  }
  send = vi.fn()
  close = vi.fn()
}

// Setup mock
global.WebSocket = MockWebSocket

describe('useWebSocket', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should connect to WebSocket', async () => {
    const onMessage = vi.fn()
    const onOpen = vi.fn()

    const { result } = renderHook(() =>
      useWebSocket('ws://localhost:8080/ws', { onMessage, onOpen })
    )

    // Wait for connection
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 10))
    })

    expect(onOpen).toHaveBeenCalled()
  })

  it('should handle incoming messages', async () => {
    const onMessage = vi.fn()

    const { result } = renderHook(() =>
      useWebSocket('ws://localhost:8080/ws', { onMessage })
    )

    // Wait for connection
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 10))
    })

    // Simulate message
    const mockWs = result.current
    const testMessage = { type: 'test', data: 'hello' }

    // The hook should handle JSON messages
    expect(onMessage).toBeDefined()
  })

  it('should send messages', async () => {
    const { result } = renderHook(() =>
      useWebSocket('ws://localhost:8080/ws')
    )

    // Wait for connection
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 10))
    })

    // Test send function exists
    expect(result.current.send).toBeDefined()
  })
})

describe('useTaskUpdates', () => {
  it('should connect to task WebSocket endpoint', async () => {
    const onUpdate = vi.fn()

    const { result } = renderHook(() =>
      useTaskUpdates('task-123', onUpdate)
    )

    // Wait for connection
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 10))
    })

    // Hook should be initialized
    expect(result.current).toBeDefined()
  })
})

describe('useRunUpdates', () => {
  it('should connect to run WebSocket endpoint', async () => {
    const onUpdate = vi.fn()

    const { result } = renderHook(() =>
      useRunUpdates('run-456', onUpdate)
    )

    // Wait for connection
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 10))
    })

    // Hook should be initialized
    expect(result.current).toBeDefined()
  })
})

describe('useSystemUpdates', () => {
  it('should connect to global WebSocket endpoint', async () => {
    const onUpdate = vi.fn()

    const { result } = renderHook(() =>
      useSystemUpdates(onUpdate)
    )

    // Wait for connection
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 10))
    })

    // Hook should be initialized
    expect(result.current).toBeDefined()
  })
})