import React, { useEffect, useState } from 'react';
import './KeyboardShortcuts.css';

interface Shortcut {
  key: string;
  description: string;
  category: string;
}

const shortcuts: Shortcut[] = [
  // Navigation
  { key: 'g h', description: '前往首页', category: '导航' },
  { key: 'g a', description: '前往 Agent 页面', category: '导航' },
  { key: 'g w', description: '前往工作流页面', category: '导航' },
  { key: 'g e', description: '前往执行页面', category: '导航' },
  { key: 'g s', description: '前往定时任务页面', category: '导航' },
  { key: '/', description: '打开全局搜索', category: '导航' },

  // Actions
  { key: 'n', description: '创建新项目 (Agent/工作流)', category: '操作' },
  { key: 'r', description: '刷新当前页面', category: '操作' },
  { key: '?', description: '显示快捷键帮助', category: '操作' },
  { key: 'Esc', description: '关闭弹窗/取消操作', category: '操作' },

  // Execution
  { key: 'Ctrl+Enter', description: '提交表单/执行', category: '执行' },
  { key: 'Ctrl+K', description: '打开命令面板', category: '执行' },
  { key: 'Ctrl+Shift+K', description: '关闭命令面板', category: '执行' },
];

interface KeyboardShortcutsProps {
  onNavigate?: (path: string) => void;
  onRefresh?: () => void;
  onSearch?: () => void;
  onCreateNew?: () => void;
  onCommandPalette?: () => void;
}

export const KeyboardShortcuts: React.FC<KeyboardShortcutsProps> = ({
  onNavigate,
  onRefresh,
  onSearch,
  onCreateNew,
  onCommandPalette,
}) => {
  const [showHelp, setShowHelp] = useState(false);
  const [keySequence, setKeySequence] = useState<string[]>([]);
  const [sequenceTimeout, setSequenceTimeout] = useState<NodeJS.Timeout | null>(null);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Ignore if user is typing in an input
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        if (e.key === 'Escape') {
          (e.target as HTMLElement).blur();
        }
        return;
      }

      // Show help
      if (e.key === '?' && !e.ctrlKey && !e.metaKey) {
        setShowHelp(true);
        return;
      }

      // Close help
      if (e.key === 'Escape') {
        setShowHelp(false);
        return;
      }

      // Global search
      if (e.key === '/' && !e.ctrlKey && !e.metaKey) {
        e.preventDefault();
        onSearch?.();
        return;
      }

      // Refresh
      if (e.key === 'r' && !e.ctrlKey && !e.metaKey) {
        e.preventDefault();
        onRefresh?.();
        return;
      }

      // Create new
      if (e.key === 'n' && !e.ctrlKey && !e.metaKey) {
        e.preventDefault();
        onCreateNew?.();
        return;
      }

      // Command palette (Ctrl+K or Cmd+K)
      if (e.key === 'k' && (e.ctrlKey || e.metaKey)) {
        e.preventDefault();
        onCommandPalette?.();
        return;
      }

      // Handle key sequences (g + key)
      if (keySequence.length === 0 && e.key === 'g') {
        setKeySequence(['g']);

        // Clear sequence after 1 second
        const timeout = setTimeout(() => {
          setKeySequence([]);
        }, 1000);
        setSequenceTimeout(timeout);
        return;
      }

      if (keySequence.length === 1 && keySequence[0] === 'g') {
        if (sequenceTimeout) {
          clearTimeout(sequenceTimeout);
        }

        const pathMap: Record<string, string> = {
          'h': '/',
          'a': '/agents',
          'w': '/workflows',
          'e': '/executions',
          's': '/schedules',
        };

        if (pathMap[e.key]) {
          e.preventDefault();
          onNavigate?.(pathMap[e.key]);
        }

        setKeySequence([]);
        return;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      if (sequenceTimeout) {
        clearTimeout(sequenceTimeout);
      }
    };
  }, [keySequence, sequenceTimeout, onNavigate, onRefresh, onSearch, onCreateNew, onCommandPalette]);

  if (!showHelp) {
    return null;
  }

  const groupedShortcuts = shortcuts.reduce((acc, shortcut) => {
    if (!acc[shortcut.category]) {
      acc[shortcut.category] = [];
    }
    acc[shortcut.category].push(shortcut);
    return acc;
  }, {} as Record<string, Shortcut[]>);

  return (
    <div className="shortcuts-overlay" onClick={() => setShowHelp(false)}>
      <div className="shortcuts-modal" onClick={(e) => e.stopPropagation()}>
        <div className="shortcuts-header">
          <h2>键盘快捷键</h2>
          <button className="close-btn" onClick={() => setShowHelp(false)}>×</button>
        </div>

        <div className="shortcuts-content">
          {Object.entries(groupedShortcuts).map(([category, items]) => (
            <div key={category} className="shortcut-category">
              <h3>{category}</h3>
              <div className="shortcut-list">
                {items.map((shortcut, index) => (
                  <div key={index} className="shortcut-item">
                    <kbd className="shortcut-key">{shortcut.key}</kbd>
                    <span className="shortcut-desc">{shortcut.description}</span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>

        <div className="shortcuts-footer">
          <span>按 <kbd>Esc</kbd> 关闭</span>
          <span className="hint">按 <kbd>?</kbd> 可随时打开此帮助</span>
        </div>
      </div>
    </div>
  );
};

export default KeyboardShortcuts;