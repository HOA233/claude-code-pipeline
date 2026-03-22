import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import './CommandPalette.css';

interface Command {
  id: string;
  title: string;
  category: string;
  shortcut?: string;
  icon: string;
  action: () => void;
}

interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
  onCreateAgent?: () => void;
  onCreateWorkflow?: () => void;
  onRefresh?: () => void;
}

export const CommandPalette: React.FC<CommandPaletteProps> = ({
  isOpen,
  onClose,
  onCreateAgent,
  onCreateWorkflow,
  onRefresh,
}) => {
  const [query, setQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();

  const commands: Command[] = [
    // Navigation
    { id: 'nav-home', title: '前往首页', category: '导航', icon: '🏠', action: () => navigate('/') },
    { id: 'nav-agents', title: '前往 Agent 页面', category: '导航', icon: '🤖', action: () => navigate('/agents') },
    { id: 'nav-workflows', title: '前往工作流页面', category: '导航', icon: '🔄', action: () => navigate('/workflows') },
    { id: 'nav-executions', title: '前往执行页面', category: '导航', icon: '📊', action: () => navigate('/executions') },
    { id: 'nav-schedules', title: '前往定时任务页面', category: '导航', icon: '⏰', action: () => navigate('/schedules') },
    { id: 'nav-webhooks', title: '前往 Webhooks 页面', category: '导航', icon: '🔔', action: () => navigate('/webhooks') },
    { id: 'nav-templates', title: '前往模板库', category: '导航', icon: '📋', action: () => navigate('/templates') },
    { id: 'nav-quotas', title: '前往配额成本页面', category: '导航', icon: '📈', action: () => navigate('/quotas') },
    { id: 'nav-audit', title: '前往审计日志', category: '导航', icon: '📜', action: () => navigate('/audit-logs') },
    { id: 'nav-diagnostics', title: '前往系统诊断', category: '导航', icon: '💊', action: () => navigate('/diagnostics') },
    { id: 'nav-settings', title: '前往系统设置', category: '导航', icon: '⚙️', action: () => navigate('/settings') },
    { id: 'nav-help', title: '前往帮助中心', category: '导航', icon: '❓', action: () => navigate('/help') },

    // Actions
    { id: 'create-agent', title: '创建新 Agent', category: '操作', icon: '➕', action: () => { navigate('/agents'); onCreateAgent?.(); } },
    { id: 'create-workflow', title: '创建新工作流', category: '操作', icon: '➕', action: () => { navigate('/workflows'); onCreateWorkflow?.(); } },
    { id: 'refresh', title: '刷新页面', category: '操作', icon: '🔄', shortcut: 'r', action: () => onRefresh?.() },

    // Quick Actions
    { id: 'cancel-all-running', title: '取消所有运行中的执行', category: '快速操作', icon: '⏹', action: () => navigate('/executions') },
    { id: 'view-logs', title: '查看执行日志', category: '快速操作', icon: '📝', action: () => navigate('/executions') },
    { id: 'export-audit', title: '导出审计日志', category: '快速操作', icon: '📤', action: () => navigate('/audit-logs') },
    { id: 'system-health', title: '查看系统健康状态', category: '快速操作', icon: '💊', action: () => navigate('/diagnostics') },
  ];

  const filteredCommands = commands.filter((cmd) => {
    if (!query) return true;
    return cmd.title.toLowerCase().includes(query.toLowerCase()) ||
           cmd.category.toLowerCase().includes(query.toLowerCase());
  });

  useEffect(() => {
    if (isOpen && inputRef.current) {
      inputRef.current.focus();
      setQuery('');
      setSelectedIndex(0);
    }
  }, [isOpen]);

  useEffect(() => {
    setSelectedIndex(0);
  }, [query]);

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setSelectedIndex((prev) => Math.min(prev + 1, filteredCommands.length - 1));
        break;
      case 'ArrowUp':
        e.preventDefault();
        setSelectedIndex((prev) => Math.max(prev - 1, 0));
        break;
      case 'Enter':
        e.preventDefault();
        if (filteredCommands[selectedIndex]) {
          filteredCommands[selectedIndex].action();
          onClose();
        }
        break;
      case 'Escape':
        e.preventDefault();
        onClose();
        break;
    }
  }, [filteredCommands, selectedIndex, onClose]);

  if (!isOpen) return null;

  const groupedCommands = filteredCommands.reduce((acc, cmd) => {
    if (!acc[cmd.category]) {
      acc[cmd.category] = [];
    }
    acc[cmd.category].push(cmd);
    return acc;
  }, {} as Record<string, Command[]>);

  return (
    <div className="command-palette-overlay" onClick={onClose}>
      <div className="command-palette" onClick={(e) => e.stopPropagation()}>
        <div className="command-input-wrapper">
          <span className="command-icon">🔍</span>
          <input
            ref={inputRef}
            type="text"
            className="command-input"
            placeholder="输入命令或搜索..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
          />
          <span className="command-hint">ESC 关闭</span>
        </div>

        <div className="command-list">
          {Object.entries(groupedCommands).map(([category, cmds]) => (
            <div key={category} className="command-group">
              <div className="command-group-title">{category}</div>
              {cmds.map((cmd) => {
                const globalIndex = filteredCommands.indexOf(cmd);
                return (
                  <div
                    key={cmd.id}
                    className={`command-item ${globalIndex === selectedIndex ? 'selected' : ''}`}
                    onClick={() => {
                      cmd.action();
                      onClose();
                    }}
                    onMouseEnter={() => setSelectedIndex(globalIndex)}
                  >
                    <span className="command-item-icon">{cmd.icon}</span>
                    <span className="command-item-title">{cmd.title}</span>
                    {cmd.shortcut && (
                      <kbd className="command-item-shortcut">{cmd.shortcut}</kbd>
                    )}
                  </div>
                );
              })}
            </div>
          ))}

          {filteredCommands.length === 0 && (
            <div className="command-empty">
              <span className="command-empty-icon">🔍</span>
              <span>未找到匹配的命令</span>
            </div>
          )}
        </div>

        <div className="command-footer">
          <span><kbd>↑</kbd><kbd>↓</kbd> 选择</span>
          <span><kbd>Enter</kbd> 执行</span>
          <span><kbd>Esc</kbd> 关闭</span>
        </div>
      </div>
    </div>
  );
};

export default CommandPalette;