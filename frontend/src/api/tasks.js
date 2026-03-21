import client from './client'

export const tasksApi = {
  // Create task
  create: (data) => client.post('/tasks', data),

  // Get task list
  getList: (params) => client.get('/tasks', { params }),

  // Get task detail
  getDetail: (taskId) => client.get(`/tasks/${taskId}`),

  // Get task result
  getResult: (taskId) => client.get(`/tasks/${taskId}/result`),

  // Cancel task
  cancel: (taskId) => client.delete(`/tasks/${taskId}`),

  // Get service status
  getStatus: () => client.get('/status'),
}

export default tasksApi