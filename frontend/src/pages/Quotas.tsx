import React, { useState, useEffect, useCallback } from 'react';
import api from '../api/client';
import './Quotas.css';

interface Quota {
  id: string;
  name: string;
  resource_type: string;
  limit: number;
  used: number;
  unit: string;
  period: 'daily' | 'weekly' | 'monthly';
  reset_at: string;
}

interface CostRecord {
  id: string;
  resource_type: string;
  resource_id: string;
  resource_name: string;
  tokens_input: number;
  tokens_output: number;
  cost: number;
  timestamp: string;
}

function Quotas() {
  const [quotas, setQuotas] = useState<Quota[]>([]);
  const [costs, setCosts] = useState<CostRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<'quotas' | 'costs'>('quotas');
  const [timeRange, setTimeRange] = useState<'today' | 'week' | 'month'>('week');
  const [summary, setSummary] = useState({
    total_cost: 0,
    total_tokens_input: 0,
    total_tokens_output: 0,
  });

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [quotasRes, costsRes] = await Promise.all([
        api.getQuotas().catch(() => ({ quotas: [], total: 0 })),
        api.getCosts(timeRange).catch(() => ({ costs: [], total: 0, summary: { total_cost: 0, total_tokens_input: 0, total_tokens_output: 0 } })),
      ]);

      setQuotas(quotasRes.quotas || []);
      setCosts(costsRes.costs || []);
      setSummary(costsRes.summary || { total_cost: 0, total_tokens_input: 0, total_tokens_output: 0 });
    } catch (error) {
      console.error('Failed to fetch quota data:', error);
    } finally {
      setLoading(false);
    }
  }, [timeRange]);

  useEffect(() => {
    fetchData();
  }, [fetchData, timeRange]);

  const getUsagePercent = (used: number, limit: number) => {
    return Math.min((used / limit) * 100, 100);
  };

  const getUsageColor = (percent: number) => {
    if (percent >= 90) return '#f5222d';
    if (percent >= 70) return '#faad14';
    return '#52c41a';
  };

  const formatNumber = (num: number) => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
    return num.toString();
  };

  const formatResetTime = (dateStr: string) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diff = date.getTime() - now.getTime();
    const days = Math.floor(diff / 86400000);
    const hours = Math.floor((diff % 86400000) / 3600000);
    if (days > 0) return `${days}天${hours}小时后重置`;
    return `${hours}小时后重置`;
  };

  const handleExportReport = () => {
    const report = {
      generated_at: new Date().toISOString(),
      time_range: timeRange,
      quotas,
      costs,
      summary,
    };
    const blob = new Blob([JSON.stringify(report, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `quota-report-${new Date().toISOString().split('T')[0]}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="quotas-page">
      <div className="page-header">
        <h1>配额与成本</h1>
        <p>监控资源使用和成本消耗</p>
      </div>

      {/* Summary Cards */}
      <div className="summary-cards">
        <div className="summary-card">
          <div className="summary-label">本月总成本</div>
          <div className="summary-value">${summary.total_cost.toFixed(3)}</div>
        </div>
        <div className="summary-card">
          <div className="summary-label">输入 Token</div>
          <div className="summary-value">{formatNumber(summary.total_tokens_input)}</div>
        </div>
        <div className="summary-card">
          <div className="summary-label">输出 Token</div>
          <div className="summary-value">{formatNumber(summary.total_tokens_output)}</div>
        </div>
        <div className="summary-card">
          <div className="summary-label">执行次数</div>
          <div className="summary-value">{costs.length}</div>
        </div>
      </div>

      {/* Tabs */}
      <div className="tabs-bar">
        <button
          className={`tab-btn ${activeTab === 'quotas' ? 'active' : ''}`}
          onClick={() => setActiveTab('quotas')}
        >
          配额管理
        </button>
        <button
          className={`tab-btn ${activeTab === 'costs' ? 'active' : ''}`}
          onClick={() => setActiveTab('costs')}
        >
          成本明细
        </button>
      </div>

      {loading ? (
        <div className="loading">加载中...</div>
      ) : (
        <>
          {activeTab === 'quotas' && (
            <div className="quotas-grid">
              {quotas.map((quota) => {
                const percent = getUsagePercent(quota.used, quota.limit);
                return (
                  <div key={quota.id} className="quota-card">
                    <div className="quota-header">
                      <h3>{quota.name}</h3>
                      <span className="quota-period">{quota.period === 'daily' ? '每日' : quota.period === 'weekly' ? '每周' : '每月'}</span>
                    </div>

                    <div className="quota-usage">
                      <div className="usage-numbers">
                        <span className="used">{formatNumber(quota.used)}</span>
                        <span className="separator">/</span>
                        <span className="limit">{formatNumber(quota.limit)} {quota.unit}</span>
                      </div>
                      <span className="usage-percent" style={{ color: getUsageColor(percent) }}>
                        {percent.toFixed(1)}%
                      </span>
                    </div>

                    <div className="usage-bar">
                      <div
                        className="usage-fill"
                        style={{
                          width: `${percent}%`,
                          backgroundColor: getUsageColor(percent),
                        }}
                      />
                    </div>

                    <div className="quota-footer">
                      <span className="reset-time">{formatResetTime(quota.reset_at)}</span>
                      <span className="remaining">
                        剩余: {formatNumber(quota.limit - quota.used)} {quota.unit}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          )}

          {activeTab === 'costs' && (
            <div className="costs-section">
              <div className="costs-toolbar">
                <select value={timeRange} onChange={(e) => setTimeRange(e.target.value as any)}>
                  <option value="today">今天</option>
                  <option value="week">本周</option>
                  <option value="month">本月</option>
                </select>
                <button className="btn-export" onClick={handleExportReport}>导出报表</button>
              </div>

              <div className="costs-table-container">
                <table className="costs-table">
                  <thead>
                    <tr>
                      <th>资源类型</th>
                      <th>资源名称</th>
                      <th>输入 Token</th>
                      <th>输出 Token</th>
                      <th>成本</th>
                      <th>时间</th>
                    </tr>
                  </thead>
                  <tbody>
                    {costs.map((cost) => (
                      <tr key={cost.id}>
                        <td>
                          <span className="resource-badge">
                            {cost.resource_type === 'agent' ? '🤖' : '🔄'} {cost.resource_type}
                          </span>
                        </td>
                        <td>{cost.resource_name}</td>
                        <td>{formatNumber(cost.tokens_input)}</td>
                        <td>{formatNumber(cost.tokens_output)}</td>
                        <td className="cost-value">${cost.cost.toFixed(3)}</td>
                        <td className="time-value">
                          {new Date(cost.timestamp).toLocaleString('zh-CN')}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}

export default Quotas;