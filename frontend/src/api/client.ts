// API Client for Claude Code Agent Orchestration Platform

import type {
  Agent,
  AgentCreateRequest,
  AgentExecuteRequest,
  Workflow,
  WorkflowCreateRequest,
  Execution,
  ExecutionCreateRequest,
  ExecutionListResponse,
  ExecutionFilter,
  ScheduledJob,
  ScheduledJobCreateRequest,
  JobExecutionHistory,
} from './types';

const API_BASE = '/api';

class APIClient {
  private baseUrl: string;
  private token?: string;

  constructor(baseUrl: string = API_BASE) {
    this.baseUrl = baseUrl;
  }

  setToken(token: string) {
    this.token = token;
  }

  private async request<T>(
    method: string,
    path: string,
    body?: any
  ): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${this.baseUrl}${path}`, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Request failed');
    }

    return response.json();
  }

  // ==================== Agent API ====================

  async createAgent(data: AgentCreateRequest): Promise<Agent> {
    return this.request('POST', '/agents', data);
  }

  async getAgent(id: string): Promise<Agent> {
    return this.request('GET', `/agents/${id}`);
  }

  async listAgents(params?: { tenant_id?: string; category?: string }): Promise<{
    agents: Agent[];
    total: number;
  }> {
    const query = new URLSearchParams();
    if (params?.tenant_id) query.set('tenant_id', params.tenant_id);
    if (params?.category) query.set('category', params.category);
    return this.request('GET', `/agents?${query}`);
  }

  async updateAgent(id: string, data: Partial<AgentCreateRequest>): Promise<Agent> {
    return this.request('PUT', `/agents/${id}`, data);
  }

  async deleteAgent(id: string): Promise<void> {
    return this.request('DELETE', `/agents/${id}`);
  }

  async testAgent(id: string, input: Record<string, any>): Promise<any> {
    return this.request('POST', `/agents/${id}/test`, input);
  }

  async executeAgent(id: string, data: AgentExecuteRequest): Promise<Execution> {
    return this.request('POST', `/agents/${id}/execute`, data);
  }

  // ==================== Workflow API ====================

  async createWorkflow(data: WorkflowCreateRequest): Promise<Workflow> {
    return this.request('POST', '/workflows', data);
  }

  async getWorkflow(id: string): Promise<Workflow> {
    return this.request('GET', `/workflows/${id}`);
  }

  async listWorkflows(params?: { tenant_id?: string }): Promise<{
    workflows: Workflow[];
    total: number;
  }> {
    const query = new URLSearchParams();
    if (params?.tenant_id) query.set('tenant_id', params.tenant_id);
    return this.request('GET', `/workflows?${query}`);
  }

  async updateWorkflow(id: string, data: Partial<WorkflowCreateRequest>): Promise<Workflow> {
    return this.request('PUT', `/workflows/${id}`, data);
  }

  async deleteWorkflow(id: string): Promise<void> {
    return this.request('DELETE', `/workflows/${id}`);
  }

  // ==================== Execution API ====================

  async executeWorkflow(data: ExecutionCreateRequest): Promise<Execution> {
    return this.request('POST', '/executions', data);
  }

  async getExecution(id: string): Promise<Execution> {
    return this.request('GET', `/executions/${id}`);
  }

  async listExecutions(filter?: ExecutionFilter): Promise<ExecutionListResponse> {
    const query = new URLSearchParams();
    if (filter?.status) query.set('status', filter.status);
    if (filter?.workflow_id) query.set('workflow_id', filter.workflow_id);
    if (filter?.tenant_id) query.set('tenant_id', filter.tenant_id);
    if (filter?.page) query.set('page', String(filter.page));
    if (filter?.page_size) query.set('page_size', String(filter.page_size));
    return this.request('GET', `/executions?${query}`);
  }

  async cancelExecution(id: string): Promise<void> {
    return this.request('POST', `/executions/${id}/cancel`);
  }

  async pauseExecution(id: string): Promise<void> {
    return this.request('POST', `/executions/${id}/pause`);
  }

  async resumeExecution(id: string): Promise<void> {
    return this.request('POST', `/executions/${id}/resume`);
  }

  async cancelAllExecutions(status?: string): Promise<{ count: number }> {
    const query = status ? `?status=${status}` : '';
    return this.request('POST', `/executions/cancel-all${query}`);
  }

  // ==================== Scheduled Job API ====================

  async createJob(data: ScheduledJobCreateRequest): Promise<ScheduledJob> {
    return this.request('POST', '/schedules', data);
  }

  async getJob(id: string): Promise<ScheduledJob> {
    return this.request('GET', `/schedules/${id}`);
  }

  async listJobs(params?: { tenant_id?: string }): Promise<{
    jobs: ScheduledJob[];
    total: number;
  }> {
    const query = new URLSearchParams();
    if (params?.tenant_id) query.set('tenant_id', params.tenant_id);
    return this.request('GET', `/schedules?${query}`);
  }

  async updateJob(id: string, data: Partial<ScheduledJobCreateRequest>): Promise<ScheduledJob> {
    return this.request('PUT', `/schedules/${id}`, data);
  }

  async deleteJob(id: string): Promise<void> {
    return this.request('DELETE', `/schedules/${id}`);
  }

  async enableJob(id: string): Promise<void> {
    return this.request('POST', `/schedules/${id}/enable`);
  }

  async disableJob(id: string): Promise<void> {
    return this.request('POST', `/schedules/${id}/disable`);
  }

  async triggerJob(id: string): Promise<Execution> {
    return this.request('POST', `/schedules/${id}/trigger`);
  }

  async getJobHistory(
    id: string,
    page = 1,
    pageSize = 20
  ): Promise<{
    history: JobExecutionHistory[];
    total: number;
    page: number;
    page_size: number;
  }> {
    return this.request('GET', `/schedules/${id}/history?page=${page}&page_size=${pageSize}`);
  }

  // ==================== SSE / WebSocket ====================

  subscribeExecution(id: string, onUpdate: (data: any) => void): () => void {
    const eventSource = new EventSource(`/sse/executions/${id}`);

    eventSource.addEventListener('execution_update', (event) => {
      const data = JSON.parse(event.data);
      onUpdate(data);
    });

    return () => eventSource.close();
  }

  subscribeAllExecutions(onUpdate: (data: any) => void): () => void {
    const eventSource = new EventSource('/sse/executions');

    eventSource.addEventListener('execution_update', (event) => {
      const data = JSON.parse(event.data);
      onUpdate(data);
    });

    return () => eventSource.close();
  }

  connectWebSocket(onMessage: (data: any) => void): WebSocket {
    const ws = new WebSocket(`ws://${window.location.host}/ws/executions`);

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      onMessage(data);
    };

    return ws;
  }
}

export const api = new APIClient();
export default api;