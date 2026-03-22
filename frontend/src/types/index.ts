// Agent Types
export interface Agent {
  id: string;
  name: string;
  description: string;
  model: string;
  system_prompt: string;
  max_tokens: number;
  skills: SkillRef[];
  default_skill?: string;
  tools: Tool[];
  permissions: Permission[];
  input_schema?: any;
  output_schema?: any;
  timeout: number;
  retry_policy?: RetryPolicy;
  isolation: IsolationConfig;
  tags: string[];
  category?: string;
  version: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
  tenant_id?: string;
}

export interface SkillRef {
  skill_id: string;
  alias?: string;
  input_mapping?: Record<string, string>;
  output_mapping?: Record<string, string>;
}

export interface Tool {
  name: string;
  description: string;
  config?: Record<string, any>;
}

export interface Permission {
  resource: string;
  action: string;
}

export interface RetryPolicy {
  max_retries: number;
  backoff: 'linear' | 'exponential';
  max_delay: number;
}

export interface IsolationConfig {
  data_isolation: boolean;
  session_isolation: boolean;
  network_isolation: boolean;
  file_isolation: boolean;
  namespace?: string;
}

export interface AgentCreateRequest {
  name: string;
  description?: string;
  model?: string;
  system_prompt?: string;
  max_tokens?: number;
  skills?: SkillRef[];
  tools?: Tool[];
  permissions?: Permission[];
  input_schema?: any;
  output_schema?: any;
  timeout?: number;
  isolation?: IsolationConfig;
  tags?: string[];
  category?: string;
}

export interface AgentExecuteRequest {
  input: Record<string, any>;
  context?: Record<string, any>;
  async?: boolean;
  callback?: string;
}

// Workflow Types
export interface Workflow {
  id: string;
  name: string;
  description: string;
  agents: AgentNode[];
  connections?: Connection[];
  mode: ExecutionMode;
  session_id?: string;
  tenant_id?: string;
  context?: Record<string, any>;
  error_handling?: ErrorConfig;
  output?: OutputConfig;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface AgentNode {
  id: string;
  name?: string;
  agent_id: string;
  input?: Record<string, any>;
  input_from?: Record<string, string>;
  output_as?: string;
  depends_on?: string[];
  condition?: string;
  timeout?: number;
  on_error?: ErrorStrategy;
  retry_count?: number;
  cli?: string;
  action?: string;
  command?: string;
  params?: Record<string, any>;
}

export interface Connection {
  from_node: string;
  from_output: string;
  to_node: string;
  to_input: string;
}

export type ExecutionMode = 'serial' | 'parallel' | 'hybrid';
export type ErrorStrategy = 'continue' | 'stop' | 'retry';

export interface ErrorConfig {
  retry: number;
  on_failure: string;
  webhook?: string;
  notify_email?: string;
}

export interface OutputConfig {
  format: string;
  merge_strategy: string;
  save_path?: string;
}

export interface WorkflowCreateRequest {
  name: string;
  description?: string;
  agents: AgentNode[];
  connections?: Connection[];
  mode?: ExecutionMode;
  context?: Record<string, any>;
  error_handling?: ErrorConfig;
  output?: OutputConfig;
}

// Execution Types
export type ExecutionStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled' | 'paused';

export interface Execution {
  id: string;
  workflow_id: string;
  workflow_name: string;
  session_id: string;
  tenant_id?: string;
  status: ExecutionStatus;
  progress: number;
  current_step?: string;
  total_steps: number;
  completed_steps: number;
  node_results: Record<string, NodeResult>;
  final_output?: any;
  duration: number;
  error?: string;
  created_at: string;
  started_at?: string;
  completed_at?: string;
  updated_at: string;
}

export interface NodeResult {
  node_id: string;
  agent_id: string;
  status: ExecutionStatus;
  output?: any;
  error?: string;
  duration: number;
  started_at?: string;
  completed_at?: string;
  retries: number;
}

export interface ExecutionCreateRequest {
  workflow_id: string;
  input?: Record<string, any>;
  context?: Record<string, any>;
  async?: boolean;
  callback?: string;
}

export interface ExecutionListResponse {
  executions: Execution[];
  total: number;
  page: number;
  page_size: number;
}

export interface ExecutionFilter {
  status?: ExecutionStatus;
  workflow_id?: string;
  tenant_id?: string;
  start_date?: string;
  end_date?: string;
  page?: number;
  page_size?: number;
}

// Scheduled Job Types
export interface ScheduledJob {
  id: string;
  name: string;
  description: string;
  target_type: 'agent' | 'workflow';
  target_id: string;
  target_name?: string;
  cron: string;
  timezone: string;
  enabled: boolean;
  input?: Record<string, any>;
  last_run?: string;
  next_run?: string;
  run_count: number;
  last_status?: string;
  on_failure: string;
  notify_email?: string;
  retry_count?: number;
  created_at: string;
  updated_at: string;
  tenant_id?: string;
  last_error?: string;
}

export interface ScheduledJobCreateRequest {
  name: string;
  description?: string;
  target_type: 'agent' | 'workflow';
  target_id: string;
  cron: string;
  timezone?: string;
  input?: Record<string, any>;
  on_failure?: string;
  notify_email?: string;
}

export interface JobExecutionHistory {
  id: string;
  job_id: string;
  execution_id: string;
  status: string;
  started_at: string;
  completed_at?: string;
  error?: string;
  duration: number;
}

// SSE Event Types
export interface SSEExecutionUpdate {
  event: string;
  data: {
    execution_id: string;
    status: ExecutionStatus;
    progress: number;
    current_step?: string;
    step_status?: Record<string, string>;
    error?: string;
  };
}

// API Response Types
export interface APIError {
  error: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

// Webhook Types
export interface Webhook {
  id: string;
  name: string;
  url: string;
  events: string[];
  enabled: boolean;
  secret?: string;
  created_at: string;
  updated_at: string;
  last_triggered?: string;
  delivery_count: number;
  tenant_id?: string;
}

export interface WebhookCreateRequest {
  name: string;
  url: string;
  events: string[];
  secret?: string;
  enabled?: boolean;
}

export interface WebhookDelivery {
  id: string;
  webhook_id: string;
  event: string;
  status: number;
  duration: number;
  timestamp: string;
  request?: any;
  response?: any;
  error?: string;
}