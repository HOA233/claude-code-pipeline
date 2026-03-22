import React, { useState, useEffect, useCallback } from 'react';
import api from '../api/client';

interface ExecutionStats {
  total: number;
  completed: number;
  failed: number;
  running: number;
  pending: number;
  success_rate: number;
  avg_duration: number;
  by_status: Record<string, number>;
  by_workflow: Record<string, number>;
}

interface TrendData {
  date: string;
  total: number;
  completed: number;
  failed: number;
}

export const ExecutionCharts: React.FC = () => {
  const [stats, setStats] = useState<ExecutionStats | null>(null);
  const [trends, setTrends] = useState<TrendData[]>([]);
  const [loading, setLoading] = useState(true);
  const [days, setDays] = useState(7);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [statsRes, trendsRes] = await Promise.all([
        fetch('/api/metrics').then((r) => r.json()),
        fetch(`/api/stats/trends?days=${days}`).then((r) => r.json()),
      ]);
      setStats(statsRes);
      setTrends(trendsRes.trends || []);
    } catch (error) {
      console.error('Failed to fetch chart data:', error);
    } finally {
      setLoading(false);
    }
  }, [days]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${Math.round(ms / 1000)}s`;
    return `${Math.round(ms / 60000)}m`;
  };

  const getMaxValue = (data: TrendData[], key: keyof TrendData) => {
    return Math.max(...data.map((d) => d[key] as number), 1);
  };

  if (loading) {
    return <div className="charts-loading">加载中...</div>;
  }

  return (
    <div className="execution-charts">
      <div className="charts-header">
        <h3>执行统计</h3>
        <select value={days} onChange={(e) => setDays(parseInt(e.target.value))}>
          <option value="7">最近 7 天</option>
          <option value="14">最近 14 天</option>
          <option value="30">最近 30 天</option>
        </select>
      </div>

      {/* Summary Stats */}
      <div className="stats-summary">
        <div className="stat-item">
          <div className="stat-value">{stats?.total || 0}</div>
          <div className="stat-label">总执行数</div>
        </div>
        <div className="stat-item success">
          <div className="stat-value">{stats?.completed || 0}</div>
          <div className="stat-label">成功</div>
        </div>
        <div className="stat-item failed">
          <div className="stat-value">{stats?.failed || 0}</div>
          <div className="stat-label">失败</div>
        </div>
        <div className="stat-item running">
          <div className="stat-value">{stats?.running || 0}</div>
          <div className="stat-label">运行中</div>
        </div>
        <div className="stat-item">
          <div className="stat-value">{stats?.success_rate?.toFixed(1) || 0}%</div>
          <div className="stat-label">成功率</div>
        </div>
        <div className="stat-item">
          <div className="stat-value">{formatDuration(stats?.avg_duration || 0)}</div>
          <div className="stat-label">平均耗时</div>
        </div>
      </div>

      {/* Status Distribution */}
      <div className="chart-section">
        <h4>状态分布</h4>
        <div className="status-bar-chart">
          {stats?.by_status &&
            Object.entries(stats.by_status).map(([status, count]) => {
              const total = stats.total || 1;
              const percentage = (count / total) * 100;
              return (
                <div key={status} className="status-bar-item">
                  <div className="status-bar-label">{status}</div>
                  <div className="status-bar-track">
                    <div
                      className={`status-bar-fill status-${status}`}
                      style={{ width: `${percentage}%` }}
                    />
                  </div>
                  <div className="status-bar-value">{count}</div>
                </div>
              );
            })}
        </div>
      </div>

      {/* Trend Chart */}
      <div className="chart-section">
        <h4>执行趋势</h4>
        <div className="trend-chart">
          <div className="trend-bars">
            {trends.map((trend, index) => {
              const maxTotal = getMaxValue(trends, 'total');
              const heightPercent = (trend.total / maxTotal) * 100;
              const completedPercent = trend.total > 0
                ? (trend.completed / trend.total) * 100
                : 0;
              const failedPercent = trend.total > 0
                ? (trend.failed / trend.total) * 100
                : 0;

              return (
                <div key={index} className="trend-bar-group">
                  <div className="trend-bar-wrapper">
                    <div
                      className="trend-bar"
                      style={{ height: `${heightPercent}%` }}
                    >
                      <div
                        className="trend-segment completed"
                        style={{ height: `${completedPercent}%` }}
                      />
                      <div
                        className="trend-segment failed"
                        style={{ height: `${failedPercent}%` }}
                      />
                    </div>
                  </div>
                  <div className="trend-label">{trend.date.slice(5)}</div>
                  <div className="trend-count">{trend.total}</div>
                </div>
              );
            })}
          </div>
          <div className="trend-legend">
            <span className="legend-item">
              <span className="legend-color completed" /> 完成
            </span>
            <span className="legend-item">
              <span className="legend-color failed" /> 失败
            </span>
          </div>
        </div>
      </div>

      {/* Workflow Distribution */}
      <div className="chart-section">
        <h4>工作流分布</h4>
        <div className="workflow-list">
          {stats?.by_workflow &&
            Object.entries(stats.by_workflow)
              .sort(([, a], [, b]) => b - a)
              .slice(0, 10)
              .map(([workflow, count]) => (
                <div key={workflow} className="workflow-item">
                  <span className="workflow-name">{workflow || '未命名'}</span>
                  <span className="workflow-count">{count}</span>
                </div>
              ))}
        </div>
      </div>
    </div>
  );
};

export default ExecutionCharts;