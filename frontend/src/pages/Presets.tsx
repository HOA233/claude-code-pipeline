import React, { useState, useEffect, useCallback } from 'react';
import api from '../api/client';
import { useToast } from '../components/Toast';
import './Presets.css';

interface Preset {
  id: string;
  name: string;
  description: string;
  category: string;
  config: {
    model?: string;
    system_prompt?: string;
    max_tokens?: number;
    timeout?: number;
    tools?: string[];
  };
  created_at: string;
  is_builtin?: boolean;
}

function Presets() {
  const [presets, setPresets] = useState<Preset[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [selectedPreset, setSelectedPreset] = useState<Preset | null>(null);
  const [filter, setFilter] = useState('');
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    category: 'general',
    model: 'claude-sonnet-4-6',
    system_prompt: '',
    max_tokens: 4096,
    timeout: 300,
    tools: [] as string[],
  });
  const { addToast } = useToast();

  const fetchPresets = useCallback(async () => {
    setLoading(true);
    try {
      // Simulate presets
      const mockPresets: Preset[] = [
        {
          id: 'preset-1',
          name: '代码审查专家',
          description: '专业的代码审查配置，适合审查代码质量',
          category: 'code-review',
          config: {
            model: 'claude-sonnet-4-6',
            system_prompt: '你是一个专业的代码审查专家...',
            max_tokens: 8192,
            timeout: 300,
            tools: ['read', 'write'],
          },
          created_at: new Date(Date.now() - 86400000).toISOString(),
          is_builtin: true,
        },
        {
          id: 'preset-2',
          name: '测试生成器',
          description: '自动生成单元测试和集成测试',
          category: 'testing',
          config: {
            model: 'claude-sonnet-4-6',
            system_prompt: '你是一个测试工程师...',
            max_tokens: 4096,
            timeout: 180,
            tools: ['read', 'write', 'execute'],
          },
          created_at: new Date(Date.now() - 172800000).toISOString(),
          is_builtin: true,
        },
        {
          id: 'preset-3',
          name: '文档助手',
          description: '生成和更新项目文档',
          category: 'documentation',
          config: {
            model: 'claude-haiku-4-5',
            system_prompt: '你是一个技术文档专家...',
            max_tokens: 2048,
            timeout: 120,
            tools: ['read', 'write'],
          },
          created_at: new Date(Date.now() - 259200000).toISOString(),
          is_builtin: false,
        },
      ];
      setPresets(mockPresets);
    } catch (error) {
      console.error('Failed to fetch presets:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchPresets();
  }, [fetchPresets]);

  const handleCreate = () => {
    if (!formData.name.trim()) {
      addToast('请输入预设名称', 'warning');
      return;
    }

    const newPreset: Preset = {
      id: `preset-${Date.now()}`,
      name: formData.name,
      description: formData.description,
      category: formData.category,
      config: {
        model: formData.model,
        system_prompt: formData.system_prompt,
        max_tokens: formData.max_tokens,
        timeout: formData.timeout,
        tools: formData.tools,
      },
      created_at: new Date().toISOString(),
    };

    setPresets([...presets, newPreset]);
    setShowCreateForm(false);
    setFormData({
      name: '',
      description: '',
      category: 'general',
      model: 'claude-sonnet-4-6',
      system_prompt: '',
      max_tokens: 4096,
      timeout: 300,
      tools: [],
    });
    addToast('预设已创建', 'success');
  };

  const handleDelete = (id: string) => {
    if (!confirm('确定要删除此预设吗？')) return;
    setPresets(presets.filter((p) => p.id !== id));
    addToast('预设已删除', 'success');
  };

  const handleDuplicate = (preset: Preset) => {
    const newPreset: Preset = {
      ...preset,
      id: `preset-${Date.now()}`,
      name: `${preset.name} (副本)`,
      is_builtin: false,
      created_at: new Date().toISOString(),
    };
    setPresets([...presets, newPreset]);
    addToast('预设已复制', 'success');
  };

  const getCategoryColor = (category: string) => {
    const colors: Record<string, string> = {
      'code-review': '#52c41a',
      'testing': '#1890ff',
      'documentation': '#722ed1',
      'security': '#f5222d',
      'general': '#8c8c8c',
    };
    return colors[category] || '#8c8c8c';
  };

  const getCategoryLabel = (category: string) => {
    const labels: Record<string, string> = {
      'code-review': '代码审查',
      'testing': '测试',
      'documentation': '文档',
      'security': '安全',
      'general': '通用',
    };
    return labels[category] || category;
  };

  const availableTools = ['read', 'write', 'execute', 'delete', 'network'];

  const filteredPresets = presets.filter((preset) => {
    if (!filter) return true;
    return (
      preset.name.toLowerCase().includes(filter.toLowerCase()) ||
      preset.category.toLowerCase().includes(filter.toLowerCase())
    );
  });

  return (
    <div className="presets-page">
      <div className="page-header">
        <h1>预设管理</h1>
        <p>管理和创建 Agent 配置预设</p>
      </div>

      <div className="toolbar">
        <input
          type="text"
          placeholder="搜索预设..."
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          className="search-input"
        />
        <button className="btn-create" onClick={() => setShowCreateForm(true)}>
          + 创建预设
        </button>
      </div>

      {loading ? (
        <div className="loading">加载中...</div>
      ) : (
        <div className="presets-grid">
          {filteredPresets.map((preset) => (
            <div
              key={preset.id}
              className={`preset-card ${preset.is_builtin ? 'builtin' : ''}`}
              onClick={() => setSelectedPreset(preset)}
            >
              <div className="preset-header">
                <span
                  className="category-tag"
                  style={{ backgroundColor: getCategoryColor(preset.category) }}
                >
                  {getCategoryLabel(preset.category)}
                </span>
                {preset.is_builtin && <span className="builtin-badge">内置</span>}
              </div>

              <h3 className="preset-name">{preset.name}</h3>
              <p className="preset-description">{preset.description}</p>

              <div className="preset-config-preview">
                <span className="config-item">
                  <span className="config-label">模型:</span>
                  <span className="config-value">{preset.config.model}</span>
                </span>
                <span className="config-item">
                  <span className="config-label">超时:</span>
                  <span className="config-value">{preset.config.timeout}s</span>
                </span>
              </div>

              <div className="preset-actions" onClick={(e) => e.stopPropagation()}>
                <button onClick={() => handleDuplicate(preset)}>复制</button>
                {!preset.is_builtin && (
                  <button className="btn-danger" onClick={() => handleDelete(preset.id)}>
                    删除
                  </button>
                )}
              </div>
            </div>
          ))}

          {filteredPresets.length === 0 && (
            <div className="empty-state">
              <span className="empty-icon">📝</span>
              <p>暂无预设</p>
            </div>
          )}
        </div>
      )}

      {/* Create Form */}
      {showCreateForm && (
        <div className="modal-overlay" onClick={() => setShowCreateForm(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>创建预设</h3>
              <button className="modal-close" onClick={() => setShowCreateForm(false)}>
                ×
              </button>
            </div>
            <div className="modal-body">
              <div className="form-group">
                <label>预设名称 *</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="输入预设名称"
                />
              </div>

              <div className="form-group">
                <label>描述</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  placeholder="描述这个预设的用途"
                  rows={2}
                />
              </div>

              <div className="form-row">
                <div className="form-group">
                  <label>分类</label>
                  <select
                    value={formData.category}
                    onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                  >
                    <option value="general">通用</option>
                    <option value="code-review">代码审查</option>
                    <option value="testing">测试</option>
                    <option value="documentation">文档</option>
                    <option value="security">安全</option>
                  </select>
                </div>

                <div className="form-group">
                  <label>模型</label>
                  <select
                    value={formData.model}
                    onChange={(e) => setFormData({ ...formData, model: e.target.value })}
                  >
                    <option value="claude-sonnet-4-6">Claude Sonnet 4.6</option>
                    <option value="claude-opus-4-6">Claude Opus 4.6</option>
                    <option value="claude-haiku-4-5">Claude Haiku 4.5</option>
                  </select>
                </div>
              </div>

              <div className="form-group">
                <label>系统提示词</label>
                <textarea
                  value={formData.system_prompt}
                  onChange={(e) => setFormData({ ...formData, system_prompt: e.target.value })}
                  placeholder="定义 Agent 的行为和角色"
                  rows={3}
                />
              </div>

              <div className="form-row">
                <div className="form-group">
                  <label>最大 Tokens</label>
                  <input
                    type="number"
                    value={formData.max_tokens}
                    onChange={(e) => setFormData({ ...formData, max_tokens: parseInt(e.target.value) || 4096 })}
                  />
                </div>

                <div className="form-group">
                  <label>超时时间 (秒)</label>
                  <input
                    type="number"
                    value={formData.timeout}
                    onChange={(e) => setFormData({ ...formData, timeout: parseInt(e.target.value) || 300 })}
                  />
                </div>
              </div>

              <div className="form-group">
                <label>可用工具</label>
                <div className="tools-grid">
                  {availableTools.map((tool) => (
                    <label key={tool} className="tool-checkbox">
                      <input
                        type="checkbox"
                        checked={formData.tools.includes(tool)}
                        onChange={(e) => {
                          if (e.target.checked) {
                            setFormData({ ...formData, tools: [...formData.tools, tool] });
                          } else {
                            setFormData({ ...formData, tools: formData.tools.filter((t) => t !== tool) });
                          }
                        }}
                      />
                      <span>{tool}</span>
                    </label>
                  ))}
                </div>
              </div>
            </div>
            <div className="modal-footer">
              <button className="btn-cancel" onClick={() => setShowCreateForm(false)}>
                取消
              </button>
              <button className="btn-primary" onClick={handleCreate}>
                创建
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Detail Modal */}
      {selectedPreset && !showCreateForm && (
        <div className="modal-overlay" onClick={() => setSelectedPreset(null)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>{selectedPreset.name}</h3>
              <button className="modal-close" onClick={() => setSelectedPreset(null)}>
                ×
              </button>
            </div>
            <div className="modal-body">
              <p className="detail-description">{selectedPreset.description}</p>

              <div className="detail-section">
                <h4>配置详情</h4>
                <div className="config-details">
                  <div className="detail-row">
                    <span className="detail-label">模型</span>
                    <span className="detail-value">{selectedPreset.config.model}</span>
                  </div>
                  <div className="detail-row">
                    <span className="detail-label">最大 Tokens</span>
                    <span className="detail-value">{selectedPreset.config.max_tokens}</span>
                  </div>
                  <div className="detail-row">
                    <span className="detail-label">超时时间</span>
                    <span className="detail-value">{selectedPreset.config.timeout}s</span>
                  </div>
                  <div className="detail-row">
                    <span className="detail-label">可用工具</span>
                    <span className="detail-value">{selectedPreset.config.tools?.join(', ') || '无'}</span>
                  </div>
                </div>
              </div>

              {selectedPreset.config.system_prompt && (
                <div className="detail-section">
                  <h4>系统提示词</h4>
                  <pre className="system-prompt">{selectedPreset.config.system_prompt}</pre>
                </div>
              )}
            </div>
            <div className="modal-footer">
              <button className="btn-secondary" onClick={() => handleDuplicate(selectedPreset)}>
                复制此预设
              </button>
              <button className="btn-primary" onClick={() => setSelectedPreset(null)}>
                关闭
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default Presets;