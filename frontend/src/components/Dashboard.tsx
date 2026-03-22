import React, { useState, useEffect, useCallback } from 'react';
import type { Execution, Agent, Workflow, ScheduledJob } from '../types';
import api from '../api/client';
import { ExecutionCharts } from './ExecutionCharts';
import { SystemHealth } from './SystemHealth';
import '../pages/Dashboard.css';

interface DashboardStats {
  totalExecutions: number;
  runningExecutions: number;
  completedExecutions: number;
  failedExecutions: number;
  totalAgents: number;
  enabledAgents: number;
  totalWorkflows: number;
  enabledWorkflows: number;
  totalJobs: number;
  enabledJobs: number;
}

export const Dashboard: React.FC = () => {
  const [stats, setStats] = useState<DashboardStats>({
    totalExecutions: 0,
    runningExecutions: 0,
    completedExecutions: 0,
    failedExecutions: 0,
    totalAgents: 0,
    enabledAgents: 0,
    totalWorkflows: 0,
    enabledWorkflows: 0,
    totalJobs: 0,
    enabledJobs: 0,
  });
  const [recentExecutions, setRecentExecutions] = useState<Execution[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchDashboardData = useCallback(async () => {
    setLoading(true);
    try {
      const [executionsRes, agentsRes, workflowsRes, jobsRes] = await Promise.all([
        api.listExecutions({ page: 1, page_size: 5 }),
        api.listAgents({}),
        api.listWorkflows({}),
        api.listJobs({}),
      ]);

      const executions = executionsRes.executions || [];
      const agents = agentsRes.agents || [];
      const workflows = workflowsRes.workflows || [];
      const jobs = jobsRes.jobs || [];

      setStats({
        totalExecutions: executionsRes.total,
        runningExecutions: executions.filter((e) => e.status === 'running').length,
        completedExecutions: executions.filter((e) => e.status === 'completed').length,
        failedExecutions: executions.filter((e) => e.status === 'failed').length,
        totalAgents: agents.length,
        enabledAgents: agents.filter((a) => a.enabled).length,
        totalWorkflows: workflows.length,
        enabledWorkflows: workflows.filter((w) => w.enabled).length,
        totalJobs: jobs.length,
        enabledJobs: jobs.filter((j) => j.enabled).length,
      });

      setRecentExecutions(executions);
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchDashboardData();

    // Subscribe to real-time updates
    const unsubscribe = api.subscribeAllExecutions(() => {
      fetchDashboardData();
    });

    return unsubscribe;
  }, [fetchDashboardData]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running':
        return '#1890ff';
      case 'completed':
        return '#52c41a';
      case 'failed':
        return '#f5222d';
      case 'paused':
        return '#faad14';
      default:
        return '#8c8c8c';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'running':
        return '▶';
      case 'completed':
        return '✓';
      case 'failed':
        return '✗';
      case 'paused':
        return '⏸';
      default:
        return '○';
    }
  };

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${Math.round(ms / 1000)}s`;
    return `${Math.round(ms / 60000)}m`;
  };

  if (loading) {
    return <div className="dashboard loading">加载中...</div>;
  }

  return (
    <div className="dashboard">
      <div className="header">
        <h1>仪表盘</h1>
        <button onClick={fetchDashboardData}>刷新</button>
      </div>

      {/* Stats Cards */}
      <div className="stats-grid">
        <div className="stat-card executions">
          <div className="stat-icon">📋</div>
          <div className="stat-content">
            <div className="stat-value">{stats.totalExecutions}</div>
            <div className="stat-label">总执行数</div>
          </div>
          <div className="stat-breakdown">
            <span style={{ color: '#1890ff' }}>运行中: {stats.runningExecutions}</span>
            <span style={{ color: '#52c41a' }}>完成: {stats.completedExecutions}</span>
            <span style={{ color: '#f5222d' }}>失败: {stats.failedExecutions}</span>
          </div>
        </div>

        <div className="stat-card agents">
          <div className="stat-icon">🤖</div>
          <div className="stat-content">
            <div className="stat-value">{stats.totalAgents}</div>
            <div className="stat-label">Agent 数量</div>
          </div>
          <div className="stat-breakdown">
            <span style={{ color: '#52c41a' }}>启用: {stats.enabledAgents}</span>
            <span style={{ color: '#8c8c8c' }}>禁用: {stats.totalAgents - stats.enabledAgents}</span>
          </div>
        </div>

        <div className="stat-card workflows">
          <div className="stat-icon">🔄</div>
          <div className="stat-content">
            <div className="stat-value">{stats.totalWorkflows}</div>
            <div className="stat-label">工作流数量</div>
          </div>
          <div className="stat-breakdown">
            <span style={{ color: '#52c41a' }}>启用: {stats.enabledWorkflows}</span>
            <span style={{ color: '#8c8c8c' }}>禁用: {stats.totalWorkflows - stats.enabledWorkflows}</span>
          </div>
        </div>

        <div className="stat-card jobs">
          <div className="stat-icon">⏰</div>
          <div className="stat-content">
            <div className="stat-value">{stats.totalJobs}</div>
            <div className="stat-label">定时任务</div>
          </div>
          <div className="stat-breakdown">
            <span style={{ color: '#52c41a' }}>启用: {stats.enabledJobs}</span>
            <span style={{ color: '#8c8c8c' }}>禁用: {stats.totalJobs - stats.enabledJobs}</span>
          </div>
        </div>
      </div>

      {/* Charts and Health Row */}
      <div className="dashboard-row">
        <div className="charts-section">
          <ExecutionCharts />
        </div>
        <div className="health-section">
          <SystemHealth />
        </div>
      </div>

      {/* Recent Executions */}
      <div className="recent-executions">
        <h2>最近执行</h2>
        {recentExecutions.length > 0 ? (
          <div className="execution-table">
            <table>
              <thead>
                <tr>
                  <th>状态</th>
                  <th>工作流</th>
                  <th>进度</th>
                  <th>耗时</th>
                  <th>创建时间</th>
                </tr>
              </thead>
              <tbody>
                {recentExecutions.map((exec) => (
                  <tr key={exec.id}>
                    <td>
                      <span
                        className="status-icon"
                        style={{ color: getStatusColor(exec.status) }}
                      >
                        {getStatusIcon(exec.status)}
                      </span>
                      <span className="status-text">{exec.status}</span>
                    </td>
                    <td>{exec.workflow_name}</td>
                    <td>
                      <div className="progress-cell">
                        <div className="progress-bar">
                          <div
                            className="progress-fill"
                            style={{ width: `${exec.progress}%` }}
                          />
                        </div>
                        <span className="progress-text">{exec.progress}%</span>
                      </div>
                    </td>
                    <td>{formatDuration(exec.duration)}</td>
                    <td>{new Date(exec.created_at).toLocaleString('zh-CN')}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="empty">暂无执行记录</div>
        )}
      </div>

      {/* Quick Actions */}
      <div className="quick-actions">
        <h2>快速操作</h2>
        <div className="actions-grid">
          <button className="action-btn" onClick={() => window.location.href = '/agents'}>
            <span className="icon">🤖</span>
            <span>管理 Agent</span>
          </button>
          <button className="action-btn" onClick={() => window.location.href = '/workflows'}>
            <span className="icon">🔄</span>
            <span>管理工作流</span>
          </button>
          <button className="action-btn" onClick={() => window.location.href = '/executions'}>
            <span className="icon">📋</span>
            <span>查看执行</span>
          </button>
          <button className="action-btn" onClick={() => window.location.href = '/schedules'}>
            <span className="icon">⏰</span>
            <span>定时任务</span>
          </button>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;