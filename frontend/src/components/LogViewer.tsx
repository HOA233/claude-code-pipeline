import React, { useState, useEffect, useRef, useCallback } from 'react';
import './LogViewer.css';

interface LogEntry {
  timestamp: string;
  level: 'info' | 'warn' | 'error' | 'debug';
  message: string;
  source?: string;
}

interface LogViewerProps {
  executionId: string;
  maxHeight?: string;
  autoScroll?: boolean;
  showFilters?: boolean;
}

export const LogViewer: React.FC<LogViewerProps> = ({
  executionId,
  maxHeight = '400px',
  autoScroll = true,
  showFilters = true,
}) => {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<string>('all');
  const [searchTerm, setSearchTerm] = useState('');
  const [connected, setConnected] = useState(false);
  const logContainerRef = useRef<HTMLDivElement>(null);
  const eventSourceRef = useRef<EventSource | null>(null);

  const scrollToBottom = useCallback(() => {
    if (autoScroll && logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
    }
  }, [autoScroll]);

  useEffect(() => {
    setLoading(true);

    // Connect to SSE for real-time logs
    const eventSource = new EventSource(`/sse/executions/${executionId}/logs`);
    eventSourceRef.current = eventSource;

    eventSource.onopen = () => {
      setConnected(true);
      setLoading(false);
    };

    eventSource.addEventListener('log', (event) => {
      try {
        const log: LogEntry = JSON.parse(event.data);
        setLogs((prev) => [...prev.slice(-500), log]); // Keep last 500 logs
      } catch (e) {
        console.error('Failed to parse log:', e);
      }
    });

    eventSource.addEventListener('execution_complete', () => {
      setConnected(false);
      eventSource.close();
    });

    eventSource.onerror = () => {
      setConnected(false);
      setLoading(false);
    };

    return () => {
      eventSource.close();
    };
  }, [executionId]);

  useEffect(() => {
    scrollToBottom();
  }, [logs, scrollToBottom]);

  const getLevelColor = (level: string) => {
    switch (level) {
      case 'error':
        return '#f5222d';
      case 'warn':
        return '#faad14';
      case 'debug':
        return '#8c8c8c';
      default:
        return '#52c41a';
    }
  };

  const getLevelBg = (level: string) => {
    switch (level) {
      case 'error':
        return 'rgba(245, 34, 45, 0.1)';
      case 'warn':
        return 'rgba(250, 173, 20, 0.1)';
      case 'debug':
        return 'rgba(140, 140, 140, 0.1)';
      default:
        return 'rgba(82, 196, 26, 0.1)';
    }
  };

  const filteredLogs = logs.filter((log) => {
    const matchesFilter = filter === 'all' || log.level === filter;
    const matchesSearch = !searchTerm ||
      log.message.toLowerCase().includes(searchTerm.toLowerCase()) ||
      log.source?.toLowerCase().includes(searchTerm.toLowerCase());
    return matchesFilter && matchesSearch;
  });

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      fractionalSecondDigits: 3,
    });
  };

  const handleClear = () => {
    setLogs([]);
  };

  const handleExport = () => {
    const content = logs
      .map((log) => `[${log.timestamp}] [${log.level.toUpperCase()}] ${log.message}`)
      .join('\n');
    const blob = new Blob([content], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `execution-${executionId}-logs.txt`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="log-viewer">
      {showFilters && (
        <div className="log-toolbar">
          <div className="log-status">
            <span className={`status-dot ${connected ? 'connected' : 'disconnected'}`} />
            <span>{connected ? '实时连接中' : '已断开'}</span>
          </div>

          <div className="log-filters">
            <select value={filter} onChange={(e) => setFilter(e.target.value)}>
              <option value="all">全部级别</option>
              <option value="debug">Debug</option>
              <option value="info">Info</option>
              <option value="warn">Warn</option>
              <option value="error">Error</option>
            </select>

            <input
              type="text"
              placeholder="搜索日志..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>

          <div className="log-actions">
            <button onClick={handleClear}>清除</button>
            <button onClick={handleExport}>导出</button>
          </div>
        </div>
      )}

      <div
        className="log-container"
        ref={logContainerRef}
        style={{ maxHeight }}
      >
        {loading ? (
          <div className="log-loading">加载日志中...</div>
        ) : filteredLogs.length === 0 ? (
          <div className="log-empty">暂无日志</div>
        ) : (
          filteredLogs.map((log, index) => (
            <div
              key={index}
              className="log-entry"
              style={{ borderLeftColor: getLevelColor(log.level) }}
            >
              <span className="log-time">{formatTimestamp(log.timestamp)}</span>
              <span
                className="log-level"
                style={{
                  backgroundColor: getLevelBg(log.level),
                  color: getLevelColor(log.level),
                }}
              >
                {log.level.toUpperCase()}
              </span>
              {log.source && <span className="log-source">[{log.source}]</span>}
              <span className="log-message">{log.message}</span>
            </div>
          ))
        )}
      </div>

      <div className="log-footer">
        <span>共 {logs.length} 条日志</span>
        {filter !== 'all' && <span> | 筛选后 {filteredLogs.length} 条</span>}
      </div>
    </div>
  );
};

export default LogViewer;