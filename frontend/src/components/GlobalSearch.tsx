import React, { useState, useEffect, useCallback, useRef } from 'react';
import api from '../api/client';
import './GlobalSearch.css';

interface SearchResult {
  type: 'agent' | 'workflow' | 'execution' | 'schedule';
  id: string;
  name: string;
  description?: string;
  status?: string;
}

interface GlobalSearchProps {
  onClose?: () => void;
}

export const GlobalSearch: React.FC<GlobalSearchProps> = ({ onClose }) => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  const search = useCallback(async (q: string) => {
    if (!q.trim()) {
      setResults([]);
      return;
    }

    setLoading(true);
    try {
      const [agentsRes, workflowsRes, executionsRes, jobsRes] = await Promise.all([
        api.listAgents({}).catch(() => ({ agents: [] })),
        api.listWorkflows({}).catch(() => ({ workflows: [] })),
        api.listExecutions({ page: 1, page_size: 20 }).catch(() => ({ executions: [] })),
        api.listJobs({}).catch(() => ({ jobs: [] })),
      ]);

      const qLower = q.toLowerCase();
      const allResults: SearchResult[] = [];

      // Search agents
      (agentsRes.agents || []).forEach((agent: any) => {
        if (
          agent.name?.toLowerCase().includes(qLower) ||
          agent.description?.toLowerCase().includes(qLower) ||
          agent.category?.toLowerCase().includes(qLower)
        ) {
          allResults.push({
            type: 'agent',
            id: agent.id,
            name: agent.name,
            description: agent.description,
            status: agent.enabled ? 'enabled' : 'disabled',
          });
        }
      });

      // Search workflows
      (workflowsRes.workflows || []).forEach((workflow: any) => {
        if (
          workflow.name?.toLowerCase().includes(qLower) ||
          workflow.description?.toLowerCase().includes(qLower)
        ) {
          allResults.push({
            type: 'workflow',
            id: workflow.id,
            name: workflow.name,
            description: workflow.description,
            status: workflow.enabled ? 'enabled' : 'disabled',
          });
        }
      });

      // Search executions
      (executionsRes.executions || []).forEach((execution: any) => {
        if (
          execution.workflow_name?.toLowerCase().includes(qLower) ||
          execution.status?.toLowerCase().includes(qLower)
        ) {
          allResults.push({
            type: 'execution',
            id: execution.id,
            name: execution.workflow_name || execution.id,
            description: `Status: ${execution.status}`,
            status: execution.status,
          });
        }
      });

      // Search schedules
      (jobsRes.jobs || []).forEach((job: any) => {
        if (
          job.name?.toLowerCase().includes(qLower) ||
          job.description?.toLowerCase().includes(qLower)
        ) {
          allResults.push({
            type: 'schedule',
            id: job.id,
            name: job.name,
            description: job.cron,
            status: job.enabled ? 'enabled' : 'disabled',
          });
        }
      });

      setResults(allResults.slice(0, 20));
    } catch (error) {
      console.error('Search failed:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const debounce = setTimeout(() => {
      search(query);
    }, 300);

    return () => clearTimeout(debounce);
  }, [query, search]);

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelectedIndex((prev) => Math.min(prev + 1, results.length - 1));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelectedIndex((prev) => Math.max(prev - 1, 0));
    } else if (e.key === 'Enter' && results[selectedIndex]) {
      handleSelect(results[selectedIndex]);
    } else if (e.key === 'Escape') {
      setIsOpen(false);
      onClose?.();
    }
  };

  const handleSelect = (result: SearchResult) => {
    const routes: Record<string, string> = {
      agent: '/agents',
      workflow: '/workflows',
      execution: '/executions',
      schedule: '/schedules',
    };
    window.location.href = routes[result.type] || '/';
    setIsOpen(false);
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'agent': return '🤖';
      case 'workflow': return '🔄';
      case 'execution': return '📊';
      case 'schedule': return '⏰';
      default: return '📄';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'enabled':
      case 'completed':
        return '#52c41a';
      case 'disabled':
      case 'failed':
        return '#f5222d';
      case 'running':
        return '#1890ff';
      default:
        return '#8c8c8c';
    }
  };

  return (
    <div className="global-search">
      <div className="search-input-wrapper">
        <span className="search-icon">🔍</span>
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={handleKeyDown}
          onFocus={() => setIsOpen(true)}
          placeholder="搜索 Agent、Workflow、Execution..."
          className="search-input"
        />
        {query && (
          <button className="clear-btn" onClick={() => setQuery('')}>
            ×
          </button>
        )}
        <kbd className="shortcut-hint">⌘K</kbd>
      </div>

      {isOpen && query && (
        <div className="search-results">
          {loading ? (
            <div className="search-loading">搜索中...</div>
          ) : results.length > 0 ? (
            <>
              <div className="results-header">
                找到 {results.length} 个结果
              </div>
              {results.map((result, index) => (
                <div
                  key={`${result.type}-${result.id}`}
                  className={`result-item ${index === selectedIndex ? 'selected' : ''}`}
                  onClick={() => handleSelect(result)}
                  onMouseEnter={() => setSelectedIndex(index)}
                >
                  <span className="result-icon">{getTypeIcon(result.type)}</span>
                  <div className="result-content">
                    <div className="result-name">{result.name}</div>
                    <div className="result-meta">
                      <span className="result-type">{result.type}</span>
                      {result.description && (
                        <span className="result-desc">{result.description}</span>
                      )}
                    </div>
                  </div>
                  {result.status && (
                    <span
                      className="result-status"
                      style={{ color: getStatusColor(result.status) }}
                    >
                      {result.status}
                    </span>
                  )}
                </div>
              ))}
              <div className="results-footer">
                <span>↑↓ 选择</span>
                <span>↵ 确认</span>
                <span>Esc 关闭</span>
              </div>
            </>
          ) : (
            <div className="no-results">
              <span className="no-results-icon">🔎</span>
              <p>未找到匹配结果</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default GlobalSearch;