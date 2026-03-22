import React, { useEffect, useState, useCallback } from 'react';
import { skillsApi } from '../api/skills';
import SkillCard from '../components/SkillCard/SkillCard';
import TaskList from '../components/TaskList/TaskList';
import CreateTask from '../components/CreateTask/CreateTask';
import { QuickStats } from '../components/QuickStats';
import { ActivityFeed } from '../components/ActivityFeed';
import type { Execution, Agent, Workflow, ScheduledJob } from '../types';
import api from '../api/client';
import { ExecutionCharts } from '../components/ExecutionCharts';
import { SystemHealth } from '../components/SystemHealth';
import { ExecutionTimeline } from '../components/ExecutionTimeline';
import { ExecutionDetailModal } from '../components/ExecutionDetailModal';
import './Dashboard.css';

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

const Dashboard = () => {
  const [skills, setSkills] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedSkill, setSelectedSkill] = useState(null);
  const [showModal, setShowModal] = useState(false);
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
  const [selectedExecutionId, setSelectedExecutionId] = useState<string | null>(null);
  const [activeView, setActiveView] = useState<'overview' | 'activity'>('overview');

  const loadSkills = useCallback(async () => {
    try {
      const data = await skillsApi.getList();
      setSkills(data.skills || []);
    } catch (err) {
      console.error('Failed to load skills:', err);
    }
  }, []);

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
    loadSkills();
    fetchDashboardData();

    // Subscribe to real-time updates
    const unsubscribe = api.subscribeAllExecutions(fetchDashboardData);

    return unsubscribe;
  }, [loadSkills, fetchDashboardData]);

  const handleSkillSelect = (skill) => {
    setSelectedSkill(skill);
    setShowModal(true);
  };

  const handleTaskCreated = (task) => {
    console.log('Task created:', task);
    fetchDashboardData();
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running': return '#1890ff';
      case 'completed': return '#52c41a';
      case 'failed': return '#f5222d';
      case 'paused': return '#faad14';
      default: return '#8c8c8c';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'running': return '▶';
      case 'completed': return '✓';
      case 'failed': return '✗';
      case 'paused': return '⏸';
      default: return '○';
    }
  };

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${Math.round(ms / 1000)}s`;
    return `${Math.round(ms / 60000)}m`;
  };

  if (loading && recentExecutions.length === 0) {
    return (
      <div className="dashboard loading">
        <div className="loading-content">加载中...</div>
      </div>
    );
  }

  return (
    <div className="dashboard">
      <div className="dashboard-header">
        <div className="header-left">
          <h1>仪表盘</h1>
          <p className="dashboard-subtitle">系统概览和实时监控</p>
        </div>
        <div className="header-right">
          <div className="view-toggle">
            <button
              className={`toggle-btn ${activeView === 'overview' ? 'active' : ''}`}
              onClick={() => setActiveView('overview')}
            >
              概览
            </button>
            <button
              className={`toggle-btn ${activeView === 'activity' ? 'active' : ''}`}
              onClick={() => setActiveView('activity')}
            >
              活动
            </button>
          </div>
          <button className="btn-refresh" onClick={fetchDashboardData}>
            刷新
          </button>
        </div>
      </div>

      {/* Quick Stats */}
      <QuickStats />

      {activeView === 'overview' ? (
        <>
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
                      <tr key={exec.id} onClick={() => setSelectedExecutionId(exec.id)}>
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

          {/* Execution Timeline */}
          <div className="dashboard-row timeline-row">
            <ExecutionTimeline
              limit={5}
              onSelect={(exec) => setSelectedExecutionId(exec.id)}
            />
          </div>
        </>
      ) : (
        <div className="activity-view">
          <ActivityFeed limit={30} showHeader={true} maxHeight="600px" />
        </div>
      )}

      {/* Skills Section */}
      <section className="dashboard-section">
        <h2 className="section-title">可用技能</h2>
        <div className="skills-grid">
          {skills.slice(0, 6).map((skill) => (
            <SkillCard
              key={skill.id}
              skill={skill}
              onSelect={handleSkillSelect}
            />
          ))}
        </div>
      </section>

      {/* Recent Tasks */}
      <section className="dashboard-section">
        <h2 className="section-title">最近任务</h2>
        <TaskList onSelectTask={(task) => console.log('View task:', task)} />
      </section>

      {showModal && selectedSkill && (
        <CreateTask
          skill={selectedSkill}
          onClose={() => setShowModal(false)}
          onSubmit={handleTaskCreated}
        />
      )}

      {selectedExecutionId && (
        <ExecutionDetailModal
          executionId={selectedExecutionId}
          onClose={() => setSelectedExecutionId(null)}
        />
      )}
    </div>
  );
};

export default Dashboard;