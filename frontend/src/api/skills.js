import client from './client'

export const skillsApi = {
  // Get all skills
  getList: () => client.get('/skills'),

  // Get skill detail
  getDetail: (skillId) => client.get(`/skills/${skillId}`),

  // Sync skills from GitLab
  sync: () => client.post('/skills/sync'),
}

export default skillsApi