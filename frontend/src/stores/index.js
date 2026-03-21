import { create } from 'zustand'

// Task store
export const useTaskStore = create((set, get) => ({
  tasks: [],
  activeTask: null,
  taskOutput: {},

  setTasks: (tasks) => set({ tasks }),

  addTask: (task) => set((state) => ({
    tasks: [task, ...state.tasks],
  })),

  updateTask: (taskId, updates) => set((state) => ({
    tasks: state.tasks.map((t) =>
      t.id === taskId ? { ...t, ...updates } : t
    ),
    activeTask: state.activeTask?.id === taskId
      ? { ...state.activeTask, ...updates }
      : state.activeTask,
  })),

  setActiveTask: (task) => set({ activeTask: task }),

  appendOutput: (taskId, output) => set((state) => ({
    taskOutput: {
      ...state.taskOutput,
      [taskId]: [...(state.taskOutput[taskId] || []), output],
    },
  })),

  clearOutput: (taskId) => set((state) => {
    const newOutput = { ...state.taskOutput }
    delete newOutput[taskId]
    return { taskOutput: newOutput }
  }),
}))

// Pipeline store
export const usePipelineStore = create((set) => ({
  pipelines: [],
  activePipeline: null,
  runs: [],

  setPipelines: (pipelines) => set({ pipelines }),

  addPipeline: (pipeline) => set((state) => ({
    pipelines: [...state.pipelines, pipeline],
  })),

  updatePipeline: (pipelineId, updates) => set((state) => ({
    pipelines: state.pipelines.map((p) =>
      p.id === pipelineId ? { ...p, ...updates } : p
    ),
  })),

  setActivePipeline: (pipeline) => set({ activePipeline: pipeline }),

  setRuns: (runs) => set({ runs }),

  addRun: (run) => set((state) => ({
    runs: [run, ...state.runs],
  })),

  updateRun: (runId, updates) => set((state) => ({
    runs: state.runs.map((r) =>
      r.id === runId ? { ...r, ...updates } : r
    ),
  })),
}))

// Skill store
export const useSkillStore = create((set) => ({
  skills: [],

  setSkills: (skills) => set({ skills }),

  addSkill: (skill) => set((state) => ({
    skills: [...state.skills, skill],
  })),

  updateSkill: (skillId, updates) => set((state) => ({
    skills: state.skills.map((s) =>
      s.id === skillId ? { ...s, ...updates } : s
    ),
  })),
}))

// System stats store
export const useStatsStore = create((set) => ({
  stats: {
    taskQueueLength: 0,
    runQueueLength: 0,
    activeTasks: 0,
    activeRuns: 0,
    completedTasks: 0,
    failedTasks: 0,
  },

  updateStats: (stats) => set((state) => ({
    stats: { ...state.stats, ...stats },
  })),
}))

// UI store
export const useUIStore = create((set) => ({
  sidebarOpen: true,
  theme: 'dark',
  notifications: [],

  toggleSidebar: () => set((state) => ({
    sidebarOpen: !state.sidebarOpen,
  })),

  setTheme: (theme) => set({ theme }),

  addNotification: (notification) => set((state) => ({
    notifications: [
      ...state.notifications,
      { id: Date.now(), ...notification },
    ],
  })),

  removeNotification: (id) => set((state) => ({
    notifications: state.notifications.filter((n) => n.id !== id),
  })),

  clearNotifications: () => set({ notifications: [] }),
}))