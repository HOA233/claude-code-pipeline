import React, { useState, useEffect, useCallback } from 'react';
import type { ScheduledJob } from '../types';
import api from '../api/client';

interface ScheduledJobListProps {
  tenantId?: string;
}

export const ScheduledJobList: React.FC<ScheduledJobListProps> = ({ tenantId }) => {
  const [jobs, setJobs] = useState<ScheduledJob[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchJobs = useCallback(async () => {
    setLoading(true);
    try {
      const result = await api.listJobs({ tenant_id: tenantId });
      setJobs(result.jobs);
    } catch (error) {
      console.error('Failed to fetch jobs:', error);
    } finally {
      setLoading(false);
    }
  }, [tenantId]);

  useEffect(() => {
    fetchJobs();
  }, [fetchJobs]);

  const handleDelete = async (id: string) => {
    if (!confirm('确定要删除此定时任务吗？')) return;
    try {
      await api.deleteJob(id);
      fetchJobs();
    } catch (error) {
      console.error('Failed to delete job:', error);
    }
  };

  const handleToggle = async (job: ScheduledJob) => {
    try {
      if (job.enabled) {
        await api.disableJob(job.id);
      } else {
        await api.enableJob(job.id);
      }
      fetchJobs();
    } catch (error) {
      console.error('Failed to toggle job:', error);
    }
  };

  const handleTrigger = async (job: ScheduledJob) => {
    try {
      const execution = await api.triggerJob(job.id);
      alert(`任务已触发，执行 ID: ${execution.id}`);
    } catch (error) {
      console.error('Failed to trigger job:', error);
      alert('触发失败');
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
        <button onClick={fetchJobs}>刷新</button>
      </div>

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
                  <span className="cron">{job.cron}</span>
                  {job.timezone && (
                    <span className="timezone">{job.timezone}</span>
                  )}
                </div>

                <div className="target-info">
                  <span className="target-type">{job.target_type}</span>
                  <span className="target-id">{job.target_id}</span>
                </div>

                <div className="run-info">
                  <div className="run-times">
                    <span>执行次数: {job.run_count}</span>
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
                    <span>上次执行: {formatTime(job.last_run)}</span>
                    <span>下次执行: {formatTime(job.next_run)}</span>
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
            <div className="empty">暂无定时任务</div>
          )}
        </div>
      )}
    </div>
  );
};

export default ScheduledJobList;