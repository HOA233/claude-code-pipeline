import { describe, it, expect, beforeEach } from 'vitest'
import { useTaskStore, usePipelineStore, useSkillStore, useStatsStore, useUIStore } from '../stores'

describe('Task Store', () => {
  beforeEach(() => {
    useTaskStore.setState({
      tasks: [],
      activeTask: null,
      taskOutput: {},
    })
  })

  it('should set tasks', () => {
    const tasks = [{ id: 'task-1', status: 'pending' }]
    useTaskStore.getState().setTasks(tasks)

    expect(useTaskStore.getState().tasks).toEqual(tasks)
  })

  it('should add a task', () => {
    const task = { id: 'task-2', status: 'running' }
    useTaskStore.getState().addTask(task)

    expect(useTaskStore.getState().tasks).toContainEqual(task)
  })

  it('should update a task', () => {
    useTaskStore.getState().setTasks([{ id: 'task-1', status: 'pending' }])
    useTaskStore.getState().updateTask('task-1', { status: 'completed' })

    expect(useTaskStore.getState().tasks[0].status).toBe('completed')
  })

  it('should append output', () => {
    useTaskStore.getState().appendOutput('task-1', 'line 1')
    useTaskStore.getState().appendOutput('task-1', 'line 2')

    expect(useTaskStore.getState().taskOutput['task-1']).toEqual(['line 1', 'line 2'])
  })

  it('should clear output', () => {
    useTaskStore.getState().appendOutput('task-1', 'line 1')
    useTaskStore.getState().clearOutput('task-1')

    expect(useTaskStore.getState().taskOutput['task-1']).toBeUndefined()
  })
})

describe('Pipeline Store', () => {
  beforeEach(() => {
    usePipelineStore.setState({
      pipelines: [],
      activePipeline: null,
      runs: [],
    })
  })

  it('should set pipelines', () => {
    const pipelines = [{ id: 'pipeline-1', name: 'Test Pipeline' }]
    usePipelineStore.getState().setPipelines(pipelines)

    expect(usePipelineStore.getState().pipelines).toEqual(pipelines)
  })

  it('should add a pipeline', () => {
    const pipeline = { id: 'pipeline-2', name: 'New Pipeline' }
    usePipelineStore.getState().addPipeline(pipeline)

    expect(usePipelineStore.getState().pipelines).toContainEqual(pipeline)
  })

  it('should add a run', () => {
    const run = { id: 'run-1', status: 'running' }
    usePipelineStore.getState().addRun(run)

    expect(usePipelineStore.getState().runs).toContainEqual(run)
  })
})

describe('Skill Store', () => {
  beforeEach(() => {
    useSkillStore.setState({ skills: [] })
  })

  it('should set skills', () => {
    const skills = [{ id: 'code-review', name: 'Code Review' }]
    useSkillStore.getState().setSkills(skills)

    expect(useSkillStore.getState().skills).toEqual(skills)
  })

  it('should add a skill', () => {
    const skill = { id: 'deploy', name: 'Deploy' }
    useSkillStore.getState().addSkill(skill)

    expect(useSkillStore.getState().skills).toContainEqual(skill)
  })
})

describe('Stats Store', () => {
  beforeEach(() => {
    useStatsStore.setState({
      stats: {
        taskQueueLength: 0,
        runQueueLength: 0,
        activeTasks: 0,
        activeRuns: 0,
        completedTasks: 0,
        failedTasks: 0,
      },
    })
  })

  it('should update stats', () => {
    useStatsStore.getState().updateStats({
      taskQueueLength: 5,
      activeTasks: 3,
    })

    expect(useStatsStore.getState().stats.taskQueueLength).toBe(5)
    expect(useStatsStore.getState().stats.activeTasks).toBe(3)
    expect(useStatsStore.getState().stats.runQueueLength).toBe(0)
  })
})

describe('UI Store', () => {
  beforeEach(() => {
    useUIStore.setState({
      sidebarOpen: true,
      theme: 'dark',
      notifications: [],
    })
  })

  it('should toggle sidebar', () => {
    useUIStore.getState().toggleSidebar()
    expect(useUIStore.getState().sidebarOpen).toBe(false)

    useUIStore.getState().toggleSidebar()
    expect(useUIStore.getState().sidebarOpen).toBe(true)
  })

  it('should add notification', () => {
    useUIStore.getState().addNotification({ message: 'Test notification', type: 'info' })

    expect(useUIStore.getState().notifications).toHaveLength(1)
    expect(useUIStore.getState().notifications[0].message).toBe('Test notification')
  })

  it('should remove notification', () => {
    useUIStore.getState().addNotification({ message: 'Test', type: 'info' })
    const id = useUIStore.getState().notifications[0].id

    useUIStore.getState().removeNotification(id)

    expect(useUIStore.getState().notifications).toHaveLength(0)
  })

  it('should clear all notifications', () => {
    useUIStore.getState().addNotification({ message: 'Test 1', type: 'info' })
    useUIStore.getState().addNotification({ message: 'Test 2', type: 'info' })

    useUIStore.getState().clearNotifications()

    expect(useUIStore.getState().notifications).toHaveLength(0)
  })
})