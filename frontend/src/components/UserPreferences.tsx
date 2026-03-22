import React, { useState, useEffect, useCallback } from 'react';
import api from '../api/client';
import { useToast } from './Toast';
import './UserPreferences.css';

interface NotificationPrefs {
  enabled: boolean;
  email: boolean;
  desktop: boolean;
  execution_done: boolean;
  errors: boolean;
  scheduled_jobs: boolean;
  sound: boolean;
}

interface DashboardPrefs {
  default_view: string;
  charts_enabled: boolean;
  timeline_limit: number;
  show_quick_stats: boolean;
}

interface Preferences {
  theme: string;
  language: string;
  sidebar_collapsed: boolean;
  auto_refresh: boolean;
  refresh_interval: number;
  notifications: NotificationPrefs;
  dashboard: DashboardPrefs;
}

interface UserPreferencesProps {
  onThemeChange?: (theme: string) => void;
}

export const UserPreferences: React.FC<UserPreferencesProps> = ({ onThemeChange }) => {
  const [prefs, setPrefs] = useState<Preferences>({
    theme: 'dark',
    language: 'zh-CN',
    sidebar_collapsed: false,
    auto_refresh: true,
    refresh_interval: 10,
    notifications: {
      enabled: true,
      email: false,
      desktop: true,
      execution_done: true,
      errors: true,
      scheduled_jobs: true,
      sound: false,
    },
    dashboard: {
      default_view: 'overview',
      charts_enabled: true,
      timeline_limit: 5,
      show_quick_stats: true,
    },
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const { addToast } = useToast();

  const fetchPrefs = useCallback(async () => {
    setLoading(true);
    try {
      const result = await api.getPreferences();
      setPrefs(result);
    } catch (error) {
      console.error('Failed to fetch preferences:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchPrefs();
  }, [fetchPrefs]);

  const handleSave = async () => {
    setSaving(true);
    try {
      await api.updatePreferences(prefs);
      addToast('设置已保存', 'success');
      onThemeChange?.(prefs.theme);
    } catch (error) {
      console.error('Failed to save preferences:', error);
      addToast('保存失败', 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleReset = async () => {
    if (!confirm('确定要重置所有设置为默认值吗？')) return;

    try {
      await api.resetPreferences();
      addToast('设置已重置', 'success');
      fetchPrefs();
    } catch (error) {
      console.error('Failed to reset preferences:', error);
      addToast('重置失败', 'error');
    }
  };

  const handleExport = async () => {
    try {
      const result = await api.exportPreferences();
      const blob = new Blob([JSON.stringify(result.preferences, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'user-preferences.json';
      a.click();
      URL.revokeObjectURL(url);
      addToast('设置已导出', 'success');
    } catch (error) {
      console.error('Failed to export preferences:', error);
      addToast('导出失败', 'error');
    }
  };

  const handleImport = () => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.json';
    input.onchange = async (e: any) => {
      const file = e.target.files[0];
      if (!file) return;

      const reader = new FileReader();
      reader.onload = async (event) => {
        try {
          const imported = JSON.parse(event.target?.result as string);
          await api.importPreferences(imported);
          setPrefs(imported);
          addToast('设置已导入', 'success');
        } catch (err) {
          addToast('导入失败：无效的配置文件', 'error');
        }
      };
      reader.readAsText(file);
    };
    input.click();
  };

  if (loading) {
    return <div className="preferences-loading">加载中...</div>;
  }

  return (
    <div className="user-preferences">
      {/* Appearance */}
      <div className="prefs-section">
        <h3>外观</h3>

        <div className="pref-item">
          <label>主题</label>
          <div className="theme-options">
            <button
              className={`theme-btn ${prefs.theme === 'light' ? 'active' : ''}`}
              onClick={() => setPrefs({ ...prefs, theme: 'light' })}
            >
              ☀️ 浅色
            </button>
            <button
              className={`theme-btn ${prefs.theme === 'dark' ? 'active' : ''}`}
              onClick={() => setPrefs({ ...prefs, theme: 'dark' })}
            >
              🌙 深色
            </button>
            <button
              className={`theme-btn ${prefs.theme === 'system' ? 'active' : ''}`}
              onClick={() => setPrefs({ ...prefs, theme: 'system' })}
            >
              💻 跟随系统
            </button>
          </div>
        </div>

        <div className="pref-item">
          <label>语言</label>
          <select
            value={prefs.language}
            onChange={(e) => setPrefs({ ...prefs, language: e.target.value })}
          >
            <option value="zh-CN">简体中文</option>
            <option value="en-US">English</option>
          </select>
        </div>

        <div className="pref-item toggle-item">
          <label>折叠侧边栏</label>
          <label className="toggle-switch">
            <input
              type="checkbox"
              checked={prefs.sidebar_collapsed}
              onChange={(e) => setPrefs({ ...prefs, sidebar_collapsed: e.target.checked })}
            />
            <span className="toggle-slider"></span>
          </label>
        </div>
      </div>

      {/* Behavior */}
      <div className="prefs-section">
        <h3>行为</h3>

        <div className="pref-item toggle-item">
          <label>自动刷新</label>
          <label className="toggle-switch">
            <input
              type="checkbox"
              checked={prefs.auto_refresh}
              onChange={(e) => setPrefs({ ...prefs, auto_refresh: e.target.checked })}
            />
            <span className="toggle-slider"></span>
          </label>
        </div>

        {prefs.auto_refresh && (
          <div className="pref-item">
            <label>刷新间隔 (秒)</label>
            <input
              type="number"
              min="5"
              max="300"
              value={prefs.refresh_interval}
              onChange={(e) => setPrefs({ ...prefs, refresh_interval: parseInt(e.target.value) || 10 })}
            />
          </div>
        )}
      </div>

      {/* Notifications */}
      <div className="prefs-section">
        <h3>通知</h3>

        <div className="pref-item toggle-item">
          <label>启用通知</label>
          <label className="toggle-switch">
            <input
              type="checkbox"
              checked={prefs.notifications.enabled}
              onChange={(e) => setPrefs({
                ...prefs,
                notifications: { ...prefs.notifications, enabled: e.target.checked }
              })}
            />
            <span className="toggle-slider"></span>
          </label>
        </div>

        {prefs.notifications.enabled && (
          <div className="notification-sub-options">
            <div className="pref-item toggle-item">
              <label>桌面通知</label>
              <label className="toggle-switch">
                <input
                  type="checkbox"
                  checked={prefs.notifications.desktop}
                  onChange={(e) => setPrefs({
                    ...prefs,
                    notifications: { ...prefs.notifications, desktop: e.target.checked }
                  })}
                />
                <span className="toggle-slider"></span>
              </label>
            </div>

            <div className="pref-item toggle-item">
              <label>执行完成通知</label>
              <label className="toggle-switch">
                <input
                  type="checkbox"
                  checked={prefs.notifications.execution_done}
                  onChange={(e) => setPrefs({
                    ...prefs,
                    notifications: { ...prefs.notifications, execution_done: e.target.checked }
                  })}
                />
                <span className="toggle-slider"></span>
              </label>
            </div>

            <div className="pref-item toggle-item">
              <label>错误通知</label>
              <label className="toggle-switch">
                <input
                  type="checkbox"
                  checked={prefs.notifications.errors}
                  onChange={(e) => setPrefs({
                    ...prefs,
                    notifications: { ...prefs.notifications, errors: e.target.checked }
                  })}
                />
                <span className="toggle-slider"></span>
              </label>
            </div>

            <div className="pref-item toggle-item">
              <label>定时任务通知</label>
              <label className="toggle-switch">
                <input
                  type="checkbox"
                  checked={prefs.notifications.scheduled_jobs}
                  onChange={(e) => setPrefs({
                    ...prefs,
                    notifications: { ...prefs.notifications, scheduled_jobs: e.target.checked }
                  })}
                />
                <span className="toggle-slider"></span>
              </label>
            </div>

            <div className="pref-item toggle-item">
              <label>声音提示</label>
              <label className="toggle-switch">
                <input
                  type="checkbox"
                  checked={prefs.notifications.sound}
                  onChange={(e) => setPrefs({
                    ...prefs,
                    notifications: { ...prefs.notifications, sound: e.target.checked }
                  })}
                />
                <span className="toggle-slider"></span>
              </label>
            </div>
          </div>
        )}
      </div>

      {/* Dashboard */}
      <div className="prefs-section">
        <h3>仪表盘</h3>

        <div className="pref-item">
          <label>默认视图</label>
          <select
            value={prefs.dashboard.default_view}
            onChange={(e) => setPrefs({
              ...prefs,
              dashboard: { ...prefs.dashboard, default_view: e.target.value }
            })}
          >
            <option value="overview">概览</option>
            <option value="executions">执行列表</option>
            <option value="activity">活动</option>
          </select>
        </div>

        <div className="pref-item toggle-item">
          <label>显示图表</label>
          <label className="toggle-switch">
            <input
              type="checkbox"
              checked={prefs.dashboard.charts_enabled}
              onChange={(e) => setPrefs({
                ...prefs,
                dashboard: { ...prefs.dashboard, charts_enabled: e.target.checked }
              })}
            />
            <span className="toggle-slider"></span>
          </label>
        </div>

        <div className="pref-item toggle-item">
          <label>显示快速统计</label>
          <label className="toggle-switch">
            <input
              type="checkbox"
              checked={prefs.dashboard.show_quick_stats}
              onChange={(e) => setPrefs({
                ...prefs,
                dashboard: { ...prefs.dashboard, show_quick_stats: e.target.checked }
              })}
            />
            <span className="toggle-slider"></span>
          </label>
        </div>

        <div className="pref-item">
          <label>时间线数量</label>
          <input
            type="number"
            min="3"
            max="20"
            value={prefs.dashboard.timeline_limit}
            onChange={(e) => setPrefs({
              ...prefs,
              dashboard: { ...prefs.dashboard, timeline_limit: parseInt(e.target.value) || 5 }
            })}
          />
        </div>
      </div>

      {/* Actions */}
      <div className="prefs-actions">
        <button className="btn-secondary" onClick={handleImport}>
          导入设置
        </button>
        <button className="btn-secondary" onClick={handleExport}>
          导出设置
        </button>
        <button className="btn-danger" onClick={handleReset}>
          重置为默认
        </button>
        <button className="btn-primary" onClick={handleSave} disabled={saving}>
          {saving ? '保存中...' : '保存设置'}
        </button>
      </div>
    </div>
  );
};

export default UserPreferences;