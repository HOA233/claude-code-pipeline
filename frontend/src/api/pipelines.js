import client from './client'

export const pipelinesApi = {
  // Create pipeline
  create: (data) => client.post('/pipelines', data),

  // Get all pipelines
  getList: () => client.get('/pipelines'),

  // Get pipeline detail
  getDetail: (pipelineId) => client.get(`/pipelines/${pipelineId}`),

  // Delete pipeline
  delete: (pipelineId) => client.delete(`/pipelines/${pipelineId}`),

  // Run pipeline
  run: (pipelineId, params) => client.post(`/pipelines/${pipelineId}/runs`, { params }),

  // Get pipeline runs
  getRuns: (pipelineId) => client.get(`/pipelines/${pipelineId}/runs`),

  // Get run detail
  getRun: (runId) => client.get(`/runs/${runId}`),

  // Cancel run
  cancelRun: (runId) => client.delete(`/runs/${runId}`),
}

export default pipelinesApi