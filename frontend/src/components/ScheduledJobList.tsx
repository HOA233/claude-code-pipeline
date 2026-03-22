import React, { useState, useEffect, useCallback } from 'react';
import type { ScheduledJob, Workflow, Agent } from '../types';
import api from '../api/client';
import { useToast } from './Toast';

interface ScheduledJobListProps {
  tenantId?: string;
}

export const ScheduledJobList: React.FC<ScheduledJobListProps> = ({ tenantId }) => {
  const [jobs, setJobs] = useState<ScheduledJob[]>([]);
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(false);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    target_type: 'workflow' as 'agent' | 'workflow',
    target_id: '',
    cron: '',
    timezone: 'Asia/Shanghai',
    on_failure: 'notify',
    notify_email: '',
  });
  const { addToast } = useToast();

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [jobsRes, workflowsRes, agentsRes] = await Promise.all([
        api.listJobs({ tenant_id: tenantId }),
        api.listWorkflows({ tenant_id: tenantId }),
        api.listAgents({ tenant_id: tenantId }),
      ]);
      setJobs(jobsRes.jobs || []);
      setWorkflows(workflowsRes.workflows || []);
      setAgents(agentsRes.agents || []);
    } catch (error) {
      console.error('Failed to fetch data:', error);
    } finally {
      setLoading(false);
    }
  }, [tenantId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleCreate = async () => {
    if (!formData.name || !formData.target_id || !formData.cron) {
      addToast('请填写必填字段', 'warning');
      return;
    }

    try {
      await api.createJob({
        name: formData.name,
        description: formData.description,
        target_type: formData.target_type,
        target_id: formData.target_id,
        cron: formData.cron,
        timezone: formData.timezone,
        on_failure: formData.on_failure,
        notify_email: formData.notify_email || undefined,
      });
      addToast('定时任务已创建', 'success');
      setShowCreateForm(false);
      setFormData({
        name: '',
        description: '',
        target_type: 'workflow',
        target_id: '',
        cron: '',
        timezone: 'Asia/Shanghai',
        on_failure: 'notify',
        notify_email: '',
      });
      fetchData();
    } catch (error) {
      console.error('Failed to create job:', error);
      addToast('创建失败', 'error');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('确定要删除此定时任务吗？')) return;
    try {
      await api.deleteJob(id);
      addToast('定时任务已删除', 'success');
      fetchData();
    } catch (error) {
      console.error('Failed to delete job:', error);
      addToast('删除失败', 'error');
    }
  };

  const handleToggle = async (job: ScheduledJob) => {
    try {
      if (job.enabled) {
        await api.disableJob(job.id);
        addToast('定时任务已禁用', 'info');
      } else {
        await api.enableJob(job.id);
        addToast('定时任务已启用', 'success');
      }
      fetchData();
    } catch (error) {
      console.error('Failed to toggle job:', error);
      addToast('操作失败', 'error');
    }
  };

  const handleTrigger = async (job: ScheduledJob) => {
    try {
      const execution = await api.triggerJob(job.id);
      addToast(`任务已触发，执行 ID: ${execution.id}`, 'success');
    } catch (error) {
      console.error('Failed to trigger job:', error);
      addToast('触发失败', 'error');
    }
  };

  const formatTime = (dateStr?: string) => {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleString('zh-CN');
  };

  const getStatusColor = (status?: string) => {
    switch (status) {
      case 'completed':
        return '#52c41a';
      case 'failed':
        return '#f5222d';
      case 'running':
        return '#1890ff';
      default:
        return '#8c8c8c';
    }
  };

  return (
    <div className="scheduled-job-list">
      <div className="header">
        <h2>定时任务</h2>
        <div className="header-actions">
          <button className="btn-create" onClick={() => setShowCreateForm(true)}>
            + 新建任务
          </button>
          <button onClick={fetchData}>刷新</button>
        </div>
      </div>

      {/* Create Form */}
      {showCreateForm && (
        <div className="create-form-overlay" onClick={() => setShowCreateForm(false)}>
          <div className="create-form" onClick={(e) => e.stopPropagation()}>
            <h3>新建定时任务</h3>
            <div className="form-group">
              <label>任务名称 *</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="输入任务名称"
              />
            </div>
            <div className="form-group">
              <label>描述</label>
              <input
                type="text"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="输入任务描述"
              />
            </div>
            <div className="form-row">
              <div className="form-group">
                <label>目标类型 *</label>
                <select
                  value={formData.target_type}
                  onChange={(e) => setFormData({ ...formData, target_type: e.target.value as 'agent' | 'workflow', target_id: '' })}
                >
                  <option value="workflow">工作流</option>
                  <option value="agent">Agent</option>
                </select>
              </div>
              <div className="form-group">
                <label>目标 *</label>
                <select
                  value={formData.target_id}
                  onChange={(e) => setFormData({ ...formData, target_id: e.target.value })}
                >
                  <option value="">选择目标</option>
                  {formData.target_type === 'workflow'
                    ? workflows.map((w) => (
                        <option key={w.id} value={w.id}>
                          {w.name}
                        </option>
                      ))
                    : agents.map((a) => (
                        <option key={a.id} value={a.id}>
                          {a.name}
                        </option>
                      ))}
                </select>
              </div>
            </div>
            <div className="form-row">
              <div className="form-group">
                <label>Cron 表达式 *</label>
                <input
                  type="text"
                  value={formData.cron}
                  onChange={(e) => setFormData({ ...formData, cron: e.target.value })}
                  placeholder="*/10 * * * *"
                />
              </div>
              <div className="form-group">
                <label>时区</label>
                <select
                  value={formData.timezone}
                  onChange={(e) => setFormData({ ...formData, timezone: e.target.value })}
                >
                  <option value="Asia/Shanghai">Asia/Shanghai</option>
                  <option value="UTC">UTC</option>
                  <option value="America/New_York">America/New_York</option>
                  <option value="Europe/London">Europe/London</option>
                </select>
              </div>
            </div>
            <div className="form-row">
              <div className="form-group">
                <label>失败处理</label>
                <select
                  value={formData.on_failure}
                  onChange={(e) => setFormData({ ...formData, on_failure: e.target.value })}
                >
                  <option value="notify">通知</option>
                  <option value="retry">重试</option>
                  <option value="disable">禁用任务</option>
                </select>
              </div>
              <div className="form-group">
                <label>通知邮箱</label>
                <input
                  type="email"
                  value={formData.notify_email}
                  onChange={(e) => setFormData({ ...formData, notify_email: e.target.value })}
                  placeholder="admin@example.com"
                />
              </div>
            </div>
            <div className="form-actions">
              <button className="btn-cancel" onClick={() => setShowCreateForm(false)}>
                取消
              </button>
              <button className="btn-submit" onClick={handleCreate}>
                创建
              </button>
            </div>
          </div>
        </div>
      )}

      {loading ? (
        <div className="loading">加载中...</div>
      ) : (
        <div className="list">
          {jobs.map((job) => (
            <div key={job.id} className={`job-item ${!job.enabled ? 'disabled' : ''}`}>
              <div className="main-info">
                <div className="title-row">
                  <h3>{job.name}</h3>
                  <span className={`status ${job.enabled ? 'enabled' : 'disabled'}`}>
                    {job.enabled ? '启用' : '禁用'}
                  </span>
                </div>

                <p className="description">{job.description}</p>

                <div className="schedule-info">
                  <span className="cron-badge">{job.cron}</span>
                  {job.timezone && (
                    <span className="timezone">{job.timezone}</span>
                  )}
                </div>

                <div className="target-info">
                  <span className="target-type-badge">{job.target_type}</span>
                  <span className="target-name">{job.target_name || job.target_id}</span>
                </div>

                <div className="run-info">
                  <div className="run-times">
                    <span className="run-count">执行: {job.run_count} 次</span>
                    {job.last_status && (
                      <span
                        className="last-status"
                        style={{ color: getStatusColor(job.last_status) }}
                      >
                        {job.last_status}
                      </span>
                    )}
                  </div>
                  <div className="times">
                    <span>上次: {formatTime(job.last_run)}</span>
                    <span>下次: {formatTime(job.next_run)}</span>
                  </div>
                </div>

                {job.last_error && (
                  <div className="error-info">
                    错误: {job.last_error}
                  </div>
                )}
              </div>

              <div className="actions">
                <button
                  className="trigger"
                  onClick={() => handleTrigger(job)}
                >
                  立即执行
                </button>
                <button onClick={() => handleToggle(job)}>
                  {job.enabled ? '禁用' : '启用'}
                </button>
                <button
                  className="delete"
                  onClick={() => handleDelete(job.id)}
                >
                  删除
                </button>
              </div>
            </div>
          ))}

          {jobs.length === 0 && (
            <div className="empty">暂无定时任务，点击上方"新建任务"创建</div>
          )}
        </div>
      )}
    </div>
  );
};

export default ScheduledJobList;