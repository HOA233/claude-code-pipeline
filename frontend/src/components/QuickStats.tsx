import React, { useState, useEffect, useCallback } from 'react';
import api from '../api/client';
import './QuickStats.css';

interface StatCard {
  key: string;
  label: string;
  value: number | string;
  icon: string;
  trend?: 'up' | 'down' | 'stable';
  trendValue?: string;
  color: string;
}

interface QuickStatsProps {
  refreshInterval?: number;
}

export const QuickStats: React.FC<QuickStatsProps> = ({
  refreshInterval = 30000,
}) => {
  const [stats, setStats] = useState<StatCard[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchStats = useCallback(async () => {
    try {
      const [execResult, agentsResult, workflowsResult, schedulesResult] = await Promise.all([
        api.listExecutions({ page: 1, page_size: 100 }),
        api.listAgents({}),
        api.listWorkflows({}),
        api.listJobs({}),
      ]);

      const executions = execResult.executions || [];
      const agents = agentsResult.agents || [];
      const workflows = workflowsResult.workflows || [];
      const schedules = schedulesResult.jobs || [];

      const running = executions.filter((e: any) => e.status === 'running').length;
      const completed = executions.filter((e: any) => e.status === 'completed').length;
      const failed = executions.filter((e: any) => e.status === 'failed').length;
      const successRate = executions.length > 0
        ? Math.round((completed / executions.length) * 100)
        : 0;

      const statsData: StatCard[] = [
        {
          key: 'executions',
          label: '总执行数',
          value: execResult.total || executions.length,
          icon: '📊',
          color: '#1890ff',
          trend: 'stable',
        },
        {
          key: 'running',
          label: '运行中',
          value: running,
          icon: '▶️',
          color: '#1890ff',
          trend: running > 5 ? 'up' : 'stable',
          trendValue: `${running} 个任务`,
        },
        {
          key: 'success',
          label: '成功率',
          value: `${successRate}%`,
          icon: '✅',
          color: '#52c41a',
          trend: successRate >= 90 ? 'up' : successRate >= 70 ? 'stable' : 'down',
        },
        {
          key: 'agents',
          label: 'Agent 数量',
          value: agents.length,
          icon: '🤖',
          color: '#722ed1',
          trend: 'stable',
        },
        {
          key: 'workflows',
          label: '工作流数量',
          value: workflows.length,
          icon: '🔄',
          color: '#13c2c2',
          trend: 'stable',
        },
        {
          key: 'schedules',
          label: '定时任务',
          value: schedules.filter((s: any) => s.enabled).length,
          icon: '⏰',
          color: '#faad14',
          trend: 'stable',
        },
      ];

      setStats(statsData);
    } catch (error) {
      console.error('Failed to fetch stats:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchStats();

    // Auto refresh
    const interval = setInterval(fetchStats, refreshInterval);

    // Subscribe to real-time updates
    const unsubscribe = api.subscribeAllExecutions(fetchStats);

    return () => {
      clearInterval(interval);
      unsubscribe();
    };
  }, [fetchStats, refreshInterval]);

  if (loading) {
    return (
      <div className="quick-stats loading">
        <div className="stats-skeleton">
          {[...Array(6)].map((_, i) => (
            <div key={i} className="skeleton-card" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="quick-stats">
      <div className="stats-header">
        <h3>快速统计</h3>
        <button className="refresh-btn" onClick={fetchStats}>
          刷新
        </button>
      </div>
      <div className="stats-grid">
        {stats.map((stat) => (
          <div
            key={stat.key}
            className="stat-card"
            style={{ '--stat-color': stat.color } as React.CSSProperties}
          >
            <div className="stat-icon" style={{ backgroundColor: `${stat.color}20` }}>
              <span>{stat.icon}</span>
            </div>
            <div className="stat-info">
              <div className="stat-value">{stat.value}</div>
              <div className="stat-label">{stat.label}</div>
            </div>
            {stat.trend && (
              <div className={`stat-trend ${stat.trend}`}>
                {stat.trend === 'up' && '↑'}
                {stat.trend === 'down' && '↓'}
                {stat.trend === 'stable' && '→'}
                {stat.trendValue && <span>{stat.trendValue}</span>}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};

export default QuickStats;