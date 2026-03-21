import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import Settings from '../pages/Settings'
import TaskPage from '../pages/TaskPage'

// Mock useNavigate
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useSearchParams: () => [new URLSearchParams(), vi.fn()],
  }
})

// Mock stores
vi.mock('../stores', () => ({
  useUIStore: () => ({
    theme: 'dark',
    setTheme: vi.fn(),
  }),
  useTaskStore: () => ({
    tasks: [],
    activeTask: null,
    setActiveTask: vi.fn(),
    updateTask: vi.fn(),
    appendOutput: vi.fn(),
  }),
}))

// Mock API
vi.mock('../api/skills', () => ({
  skillsApi: {
    getDetail: vi.fn().mockResolvedValue({
      id: 'code-review',
      name: 'Code Review',
      parameters: [
        { name: 'target', type: 'string', required: true },
        { name: 'depth', type: 'enum', values: ['quick', 'standard', 'deep'], default: 'standard' },
      ],
    }),
  },
}))

vi.mock('../api/tasks', () => ({
  tasksApi: {
    create: vi.fn().mockResolvedValue({ id: 'task-123', status: 'pending' }),
  },
}))

const renderWithRouter = (component) => {
  return render(
    <BrowserRouter>
      {component}
    </BrowserRouter>
  )
}

describe('Settings Page', () => {
  beforeEach(() => {
    localStorage.clear()
    mockNavigate.mockClear()
  })

  it('should render settings page', () => {
    renderWithRouter(<Settings />)

    expect(screen.getByText('Settings')).toBeInTheDocument()
    expect(screen.getByText('Appearance')).toBeInTheDocument()
  })

  it('should show API URL input', () => {
    renderWithRouter(<Settings />)

    const input = screen.getByDisplayValue('http://localhost:8080')
    expect(input).toBeInTheDocument()
  })

  it('should have theme toggle', () => {
    renderWithRouter(<Settings />)

    expect(screen.getByText('Theme')).toBeInTheDocument()
  })

  it('should show save button', () => {
    renderWithRouter(<Settings />)

    expect(screen.getByText('Save Changes')).toBeInTheDocument()
  })

  it('should save settings to localStorage', () => {
    renderWithRouter(<Settings />)

    const saveButton = screen.getByText('Save Changes')
    fireEvent.click(saveButton)

    expect(screen.getByText('✓ Saved')).toBeInTheDocument()
  })

  it('should show export button', () => {
    renderWithRouter(<Settings />)

    expect(screen.getByText('Export')).toBeInTheDocument()
  })

  it('should show clear all button', () => {
    renderWithRouter(<Settings />)

    expect(screen.getByText('Clear All')).toBeInTheDocument()
  })
})

describe('TaskPage', () => {
  beforeEach(() => {
    localStorage.clear()
    mockNavigate.mockClear()
  })

  it('should render task page', () => {
    renderWithRouter(<TaskPage />)

    expect(screen.getByText('Create Task')).toBeInTheDocument()
  })

  it('should show skill selector when no skill is selected', () => {
    renderWithRouter(<TaskPage />)

    expect(screen.getByText('Select a Skill')).toBeInTheDocument()
  })

  it('should have browse skills button', () => {
    renderWithRouter(<TaskPage />)

    expect(screen.getByText('Browse Skills')).toBeInTheDocument()
  })
})