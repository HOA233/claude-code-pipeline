import React, { useState, useEffect, useCallback } from 'react';
import api from '../api/client';
import { useToast } from '../components/Toast';
import { APIKeys } from '../components/APIKeys';
import { UserPreferences } from '../components/UserPreferences';
import './Settings.css';

interface SystemConfig {
  max_concurrent_executions: number;
  default_timeout: number;
  log_level: string;
  enable_notifications: boolean;
  notification_email: string;
  webhook_url: string;
}

interface FeatureFlags {
  [key: string]: boolean;
}

interface Model {
  id: string;
  name: string;
  provider: string;
  max_tokens: number;
  enabled: boolean;
}

function Settings() {
  const [activeTab, setActiveTab] = useState('general');
  const [config, setConfig] = useState<SystemConfig>({
    max_concurrent_executions: 10,
    default_timeout: 300,
    log_level: 'info',
    enable_notifications: false,
    notification_email: '',
    webhook_url: '',
  });
  const [features, setFeatures] = useState<FeatureFlags>({});
  const [models, setModels] = useState<Model[]>([]);
  const [apiUrl, setApiUrl] = useState(localStorage.getItem('apiUrl') || '');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const { addToast } = useToast();

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [configRes, featuresRes, modelsRes] = await Promise.all([
        api.getConfig().catch(() => null),
        api.getFeatures().catch(() => ({})),
        api.getModels().catch(() => ({ models: [] })),
      ]);
      if (configRes) setConfig(configRes);
      if (featuresRes) setFeatures(featuresRes.features || featuresRes || {});
      if (modelsRes) setModels(modelsRes.models || []);
    } catch (error) {
      console.error('Failed to fetch settings:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleSaveConfig = async () => {
    setSaving(true);
    try {
      await api.updateConfig(config);
      addToast('配置已保存', 'success');
    } catch (error) {
      console.error('Failed to save config:', error);
      addToast('保存失败', 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleToggleFeature = async (feature: string) => {
    try {
      await api.toggleFeature(feature);
      setFeatures((prev) => ({ ...prev, [feature]: !prev[feature] }));
      addToast(`功能 "${feature}" 已${features[feature] ? '禁用' : '启用'}`, 'success');
    } catch (error) {
      console.error('Failed to toggle feature:', error);
      addToast('操作失败', 'error');
    }
  };

  const handleSaveApiUrl = () => {
    localStorage.setItem('apiUrl', apiUrl);
    addToast('API URL 已保存', 'success');
  };

  const handleClearCache = () => {
    if (confirm('确定要清除所有本地数据吗？')) {
      localStorage.clear();
      window.location.reload();
    }
  };

  const handleExportSettings = () => {
    const settings = {
      config,
      features,
      apiUrl,
      exportedAt: new Date().toISOString(),
    };
    const blob = new Blob([JSON.stringify(settings, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'claude-agent-platform-settings.json';
    a.click();
    URL.revokeObjectURL(url);
  };

  if (loading) {
    return (
      <div className="settings-page loading">
        <div className="page-header">
          <h1>系统设置</h1>
        </div>
        <div className="loading-content">加载中...</div>
      </div>
    );
  }

  return (
    <div className="settings-page">
      <div className="page-header">
        <h1>系统设置</h1>
        <p>配置系统参数和功能</p>
      </div>

      <div className="tabs-bar">
        <button
          className={`tab-btn ${activeTab === 'general' ? 'active' : ''}`}
          onClick={() => setActiveTab('general')}
        >
          常规设置
        </button>
        <button
          className={`tab-btn ${activeTab === 'preferences' ? 'active' : ''}`}
          onClick={() => setActiveTab('preferences')}
        >
          个人偏好
        </button>
        <button
          className={`tab-btn ${activeTab === 'features' ? 'active' : ''}`}
          onClick={() => setActiveTab('features')}
        >
          功能开关
        </button>
        <button
          className={`tab-btn ${activeTab === 'models' ? 'active' : ''}`}
          onClick={() => setActiveTab('models')}
        >
          模型管理
        </button>
        <button
          className={`tab-btn ${activeTab === 'apikeys' ? 'active' : ''}`}
          onClick={() => setActiveTab('apikeys')}
        >
          API 密钥
        </button>
        <button
          className={`tab-btn ${activeTab === 'about' ? 'active' : ''}`}
          onClick={() => setActiveTab('about')}
        >
          关于
        </button>
      </div>

      <div className="settings-content">
        {activeTab === 'general' && (
          <div className="settings-section">
            <h3>执行配置</h3>
            <div className="form-group">
              <label>最大并发执行数</label>
              <input
                type="number"
                value={config.max_concurrent_executions}
                onChange={(e) =>
                  setConfig({ ...config, max_concurrent_executions: parseInt(e.target.value) || 10 })
                }
              />
              <span className="hint">同时执行的最大工作流数量</span>
            </div>

            <div className="form-group">
              <label>默认超时时间 (秒)</label>
              <input
                type="number"
                value={config.default_timeout}
                onChange={(e) =>
                  setConfig({ ...config, default_timeout: parseInt(e.target.value) || 300 })
                }
              />
              <span className="hint">Agent 执行的默认超时时间</span>
            </div>

            <div className="form-group">
              <label>日志级别</label>
              <select
                value={config.log_level}
                onChange={(e) => setConfig({ ...config, log_level: e.target.value })}
              >
                <option value="debug">Debug</option>
                <option value="info">Info</option>
                <option value="warn">Warn</option>
                <option value="error">Error</option>
              </select>
            </div>

            <h3>通知设置</h3>

            <div className="form-group checkbox-group">
              <label>
                <input
                  type="checkbox"
                  checked={config.enable_notifications}
                  onChange={(e) =>
                    setConfig({ ...config, enable_notifications: e.target.checked })
                  }
                />
                启用通知
              </label>
            </div>

            {config.enable_notifications && (
              <div className="form-group">
                <label>通知邮箱</label>
                <input
                  type="email"
                  value={config.notification_email}
                  onChange={(e) =>
                    setConfig({ ...config, notification_email: e.target.value })
                  }
                  placeholder="admin@example.com"
                />
              </div>
            )}

            <div className="form-group">
              <label>Webhook URL</label>
              <input
                type="url"
                value={config.webhook_url}
                onChange={(e) => setConfig({ ...config, webhook_url: e.target.value })}
                placeholder="https://webhook.example.com/notify"
              />
              <span className="hint">执行完成后的回调地址</span>
            </div>

            <h3>连接设置</h3>

            <div className="form-group">
              <label>API URL</label>
              <input
                type="text"
                value={apiUrl}
                onChange={(e) => setApiUrl(e.target.value)}
                placeholder="http://localhost:8080"
              />
              <span className="hint">API 服务器地址</span>
            </div>

            <div className="form-actions">
              <button className="btn-secondary" onClick={handleSaveApiUrl}>
                保存 API URL
              </button>
              <button className="btn-save" onClick={handleSaveConfig} disabled={saving}>
                {saving ? '保存中...' : '保存配置'}
              </button>
            </div>
          </div>
        )}

        {activeTab === 'features' && (
          <div className="settings-section">
            <h3>功能开关</h3>
            <p className="section-desc">启用或禁用平台功能</p>

            <div className="feature-list">
              {Object.entries(features).length > 0 ? (
                Object.entries(features).map(([feature, enabled]) => (
                  <div key={feature} className="feature-item">
                    <div className="feature-info">
                      <span className="feature-name">{feature}</span>
                      <span className="feature-status">{enabled ? '已启用' : '已禁用'}</span>
                    </div>
                    <label className="toggle-switch">
                      <input
                        type="checkbox"
                        checked={enabled}
                        onChange={() => handleToggleFeature(feature)}
                      />
                      <span className="toggle-slider"></span>
                    </label>
                  </div>
                ))
              ) : (
                <div className="empty-features">
                  <p>暂无可配置的功能开关</p>
                  <p className="hint">功能开关可通过 API 动态添加</p>
                </div>
              )}
            </div>
          </div>
        )}

        {activeTab === 'models' && (
          <div className="settings-section">
            <h3>可用模型</h3>
            <p className="section-desc">支持的 AI 模型列表</p>

            <div className="models-grid">
              {models.length > 0 ? (
                models.map((model) => (
                  <div key={model.id} className={`model-card ${model.enabled ? 'enabled' : 'disabled'}`}>
                    <div className="model-header">
                      <span className="model-name">{model.name}</span>
                      <span className={`model-status ${model.enabled ? 'enabled' : 'disabled'}`}>
                        {model.enabled ? '可用' : '不可用'}
                      </span>
                    </div>
                    <div className="model-details">
                      <span className="model-provider">{model.provider}</span>
                      <span className="model-tokens">最大 {model.max_tokens?.toLocaleString() || 'N/A'} tokens</span>
                    </div>
                    <div className="model-id">{model.id}</div>
                  </div>
                ))
              ) : (
                <div className="empty-models">
                  <p>暂无可用模型</p>
                  <p className="hint">模型列表可通过 API 配置</p>
                </div>
              )}
            </div>
          </div>
        )}

        {activeTab === 'apikeys' && (
          <div className="settings-section">
            <APIKeys />
          </div>
        )}

        {activeTab === 'preferences' && (
          <div className="settings-section">
            <h3>个人偏好</h3>
            <p className="section-desc">自定义您的使用体验</p>
            <UserPreferences />
          </div>
        )}

        {activeTab === 'about' && (
          <div className="settings-section">
            <h3>关于</h3>

            <div className="about-info">
              <div className="about-item">
                <span className="about-label">版本</span>
                <span className="about-value">1.0.0</span>
              </div>
              <div className="about-item">
                <span className="about-label">仓库</span>
                <a href="https://github.com/HOA233/claude-code-pipeline" target="_blank" rel="noopener noreferrer">
                  github.com/HOA233/claude-code-pipeline
                </a>
              </div>
              <div className="about-item">
                <span className="about-label">许可证</span>
                <span className="about-value">MIT</span>
              </div>
            </div>

            <h3>数据管理</h3>

            <div className="data-actions">
              <button className="btn-secondary" onClick={handleExportSettings}>
                导出设置
              </button>
              <button className="btn-danger" onClick={handleClearCache}>
                清除本地数据
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export default Settings;