import React, { useState, useEffect, useCallback } from 'react';
import api from '../api/client';
import { useToast } from '../components/Toast';
import './Templates.css';

interface TemplateAgent {
  id: string;
  name: string;
  description: string;
  model: string;
  system_prompt: string;
  timeout: number;
  output_as: string;
  depends_on: string[];
}

interface WorkflowTemplate {
  id: string;
  name: string;
  description: string;
  category: string;
  icon: string;
  agents: TemplateAgent[];
  connections: any[];
  config: Record<string, any>;
  created_at: string;
  is_builtin?: boolean;
}

function Templates() {
  const [builtinTemplates, setBuiltinTemplates] = useState<WorkflowTemplate[]>([]);
  const [customTemplates, setCustomTemplates] = useState<WorkflowTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedTemplate, setSelectedTemplate] = useState<WorkflowTemplate | null>(null);
  const [showInstantiateModal, setShowInstantiateModal] = useState(false);
  const [newInstanceName, setNewInstanceName] = useState('');
  const [activeTab, setActiveTab] = useState<'builtin' | 'custom'>('builtin');
  const { addToast } = useToast();

  const fetchTemplates = useCallback(async () => {
    setLoading(true);
    try {
      const [builtinRes, customRes] = await Promise.all([
        api.getBuiltinTemplates(),
        api.getCustomTemplates(),
      ]);
      setBuiltinTemplates((builtinRes.templates || builtinRes || []).map((t: any) => ({ ...t, is_builtin: true })));
      setCustomTemplates(customRes.templates || customRes || []);
    } catch (error) {
      console.error('Failed to fetch templates:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTemplates();
  }, [fetchTemplates]);

  const handleInstantiate = async () => {
    if (!selectedTemplate || !newInstanceName.trim()) {
      addToast('请输入工作流名称', 'warning');
      return;
    }

    try {
      const result = await api.instantiateTemplate(selectedTemplate.id, newInstanceName);
      addToast(`工作流 "${newInstanceName}" 已创建`, 'success');
      setShowInstantiateModal(false);
      setNewInstanceName('');
      setSelectedTemplate(null);
    } catch (error) {
      console.error('Failed to instantiate template:', error);
      addToast('创建失败', 'error');
    }
  };

  const handleDeleteCustom = async (id: string) => {
    if (!confirm('确定要删除此自定义模板吗？')) return;
    try {
      await api.deleteCustomTemplate(id);
      addToast('模板已删除', 'success');
      fetchTemplates();
    } catch (error) {
      console.error('Failed to delete template:', error);
      addToast('删除失败', 'error');
    }
  };

  const getCategoryColor = (category: string) => {
    const colors: Record<string, string> = {
      'code-review': '#52c41a',
      'testing': '#1890ff',
      'security': '#f5222d',
      'documentation': '#722ed1',
      'refactoring': '#faad14',
      'deployment': '#13c2c2',
      'analysis': '#eb2f96',
    };
    return colors[category] || '#8c8c8c';
  };

  const getCategoryLabel = (category: string) => {
    const labels: Record<string, string> = {
      'code-review': '代码审查',
      'testing': '测试',
      'security': '安全',
      'documentation': '文档',
      'refactoring': '重构',
      'deployment': '部署',
      'analysis': '分析',
    };
    return labels[category] || category;
  };

  const templates = activeTab === 'builtin' ? builtinTemplates : customTemplates;

  return (
    <div className="templates-page">
      <div className="page-header">
        <h1>工作流模板</h1>
        <p>从模板快速创建工作流</p>
      </div>

      <div className="tabs-bar">
        <button
          className={`tab-btn ${activeTab === 'builtin' ? 'active' : ''}`}
          onClick={() => setActiveTab('builtin')}
        >
          内置模板
        </button>
        <button
          className={`tab-btn ${activeTab === 'custom' ? 'active' : ''}`}
          onClick={() => setActiveTab('custom')}
        >
          自定义模板
        </button>
      </div>

      {loading ? (
        <div className="loading">加载中...</div>
      ) : (
        <div className="templates-grid">
          {templates.map((template) => (
            <div key={template.id} className="template-card">
              <div className="template-header">
                <span className="template-icon">{template.icon || '📋'}</span>
                <span
                  className="category-badge"
                  style={{ backgroundColor: getCategoryColor(template.category) }}
                >
                  {getCategoryLabel(template.category)}
                </span>
              </div>

              <h3 className="template-name">{template.name}</h3>
              <p className="template-description">{template.description}</p>

              <div className="template-agents">
                <div className="agents-label">
                  包含 {template.agents?.length || 0} 个 Agent
                </div>
                <div className="agents-preview">
                  {template.agents?.slice(0, 3).map((agent) => (
                    <span key={agent.id} className="agent-tag">
                      {agent.name}
                    </span>
                  ))}
                  {template.agents?.length > 3 && (
                    <span className="agent-tag more">+{template.agents.length - 3}</span>
                  )}
                </div>
              </div>

              <div className="template-actions">
                <button
                  className="btn-primary"
                  onClick={() => {
                    setSelectedTemplate(template);
                    setNewInstanceName(`${template.name} - ${new Date().toLocaleDateString()}`);
                    setShowInstantiateModal(true);
                  }}
                >
                  使用模板
                </button>
                <button
                  className="btn-secondary"
                  onClick={() => setSelectedTemplate(template)}
                >
                  查看详情
                </button>
                {!template.is_builtin && (
                  <button
                    className="btn-danger"
                    onClick={() => handleDeleteCustom(template.id)}
                  >
                    删除
                  </button>
                )}
              </div>
            </div>
          ))}

          {templates.length === 0 && (
            <div className="empty-state">
              <div className="empty-icon">📋</div>
              <p>暂无{activeTab === 'builtin' ? '内置' : '自定义'}模板</p>
            </div>
          )}
        </div>
      )}

      {/* Detail Modal */}
      {selectedTemplate && !showInstantiateModal && (
        <div className="modal-overlay" onClick={() => setSelectedTemplate(null)}>
          <div className="modal-content template-detail" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>{selectedTemplate.name}</h3>
              <button className="modal-close" onClick={() => setSelectedTemplate(null)}>
                ×
              </button>
            </div>
            <div className="modal-body">
              <p className="detail-description">{selectedTemplate.description}</p>

              <h4>Agent 流程</h4>
              <div className="agent-flow">
                {selectedTemplate.agents?.map((agent, index) => (
                  <div key={agent.id} className="flow-node">
                    <div className="node-index">{index + 1}</div>
                    <div className="node-content">
                      <div className="node-name">{agent.name}</div>
                      <div className="node-model">{agent.model}</div>
                      <div className="node-prompt">{agent.system_prompt?.slice(0, 100)}...</div>
                    </div>
                    {agent.depends_on?.length > 0 && (
                      <div className="node-deps">
                        依赖: {agent.depends_on.join(', ')}
                      </div>
                    )}
                  </div>
                ))}
              </div>

              <div className="detail-actions">
                <button
                  className="btn-primary"
                  onClick={() => {
                    setNewInstanceName(`${selectedTemplate.name} - ${new Date().toLocaleDateString()}`);
                    setShowInstantiateModal(true);
                  }}
                >
                  使用此模板
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Instantiate Modal */}
      {showInstantiateModal && selectedTemplate && (
        <div className="modal-overlay" onClick={() => setShowInstantiateModal(false)}>
          <div className="modal-content instantiate-modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>创建工作流</h3>
              <button className="modal-close" onClick={() => setShowInstantiateModal(false)}>
                ×
              </button>
            </div>
            <div className="modal-body">
              <p className="source-template">
                从模板: <strong>{selectedTemplate.name}</strong>
              </p>

              <div className="form-group">
                <label>工作流名称 *</label>
                <input
                  type="text"
                  value={newInstanceName}
                  onChange={(e) => setNewInstanceName(e.target.value)}
                  placeholder="输入工作流名称"
                />
              </div>

              <div className="form-group">
                <label>包含 Agent</label>
                <div className="agents-list">
                  {selectedTemplate.agents?.map((agent) => (
                    <span key={agent.id} className="agent-chip">
                      {agent.name}
                    </span>
                  ))}
                </div>
              </div>

              <div className="modal-actions">
                <button className="btn-cancel" onClick={() => setShowInstantiateModal(false)}>
                  取消
                </button>
                <button className="btn-primary" onClick={handleInstantiate}>
                  创建工作流
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default Templates;