import React, { useState } from 'react';
import api from '../api/client';
import type { Agent, AgentCreateRequest, Tool, Permission, SkillRef } from '../types';

interface AgentFormProps {
  agent?: Agent;
  onSuccess?: (agent: Agent) => void;
  onCancel?: () => void;
}

export const AgentForm: React.FC<AgentFormProps> = ({ agent, onSuccess, onCancel }) => {
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState<AgentCreateRequest>({
    name: agent?.name || '',
    description: agent?.description || '',
    model: agent?.model || 'claude-sonnet-4-6',
    system_prompt: agent?.system_prompt || '',
    max_tokens: agent?.max_tokens || 4096,
    timeout: agent?.timeout || 300,
    tools: agent?.tools || [],
    permissions: agent?.permissions || [],
    skills: agent?.skills || [],
    tags: agent?.tags || [],
    category: agent?.category || 'general',
  });

  const [newTool, setNewTool] = useState<Tool>({ name: '', description: '' });
  const [newPermission, setNewPermission] = useState<Permission>({ resource: '', action: '' });
  const [newTag, setNewTag] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      let result: Agent;
      if (agent?.id) {
        result = await api.updateAgent(agent.id, formData);
      } else {
        result = await api.createAgent(formData);
      }
      onSuccess?.(result);
    } catch (error) {
      console.error('Failed to save agent:', error);
      alert('保存失败');
    } finally {
      setLoading(false);
    }
  };

  const addTool = () => {
    if (newTool.name) {
      setFormData({
        ...formData,
        tools: [...(formData.tools || []), newTool],
      });
      setNewTool({ name: '', description: '' });
    }
  };

  const removeTool = (index: number) => {
    const tools = [...(formData.tools || [])];
    tools.splice(index, 1);
    setFormData({ ...formData, tools });
  };

  const addPermission = () => {
    if (newPermission.resource && newPermission.action) {
      setFormData({
        ...formData,
        permissions: [...(formData.permissions || []), newPermission],
      });
      setNewPermission({ resource: '', action: '' });
    }
  };

  const removePermission = (index: number) => {
    const permissions = [...(formData.permissions || [])];
    permissions.splice(index, 1);
    setFormData({ ...formData, permissions });
  };

  const addTag = () => {
    if (newTag && !formData.tags?.includes(newTag)) {
      setFormData({
        ...formData,
        tags: [...(formData.tags || []), newTag],
      });
      setNewTag('');
    }
  };

  const removeTag = (tag: string) => {
    setFormData({
      ...formData,
      tags: formData.tags?.filter((t) => t !== tag) || [],
    });
  };

  return (
    <form className="agent-form" onSubmit={handleSubmit}>
      <div className="form-section">
        <h3>基本信息</h3>
        <div className="form-group">
          <label>名称 *</label>
          <input
            type="text"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            required
            placeholder="Agent 名称"
          />
        </div>

        <div className="form-group">
          <label>描述</label>
          <textarea
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            placeholder="Agent 功能描述"
            rows={3}
          />
        </div>

        <div className="form-row">
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

          <div className="form-group">
            <label>分类</label>
            <select
              value={formData.category}
              onChange={(e) => setFormData({ ...formData, category: e.target.value })}
            >
              <option value="general">通用</option>
              <option value="code-review">代码审查</option>
              <option value="testing">测试</option>
              <option value="security">安全</option>
              <option value="documentation">文档</option>
              <option value="refactoring">重构</option>
            </select>
          </div>
        </div>
      </div>

      <div className="form-section">
        <h3>系统提示词</h3>
        <div className="form-group">
          <label>System Prompt</label>
          <textarea
            value={formData.system_prompt}
            onChange={(e) => setFormData({ ...formData, system_prompt: e.target.value })}
            placeholder="定义 Agent 的行为和角色..."
            rows={6}
          />
        </div>
      </div>

      <div className="form-section">
        <h3>配置</h3>
        <div className="form-row">
          <div className="form-group">
            <label>Max Tokens</label>
            <input
              type="number"
              value={formData.max_tokens}
              onChange={(e) =>
                setFormData({ ...formData, max_tokens: parseInt(e.target.value) })
              }
              min={1}
              max={100000}
            />
          </div>

          <div className="form-group">
            <label>超时时间 (秒)</label>
            <input
              type="number"
              value={formData.timeout}
              onChange={(e) =>
                setFormData({ ...formData, timeout: parseInt(e.target.value) })
              }
              min={10}
              max={3600}
            />
          </div>
        </div>
      </div>

      <div className="form-section">
        <h3>工具配置</h3>
        <div className="inline-form">
          <input
            type="text"
            placeholder="工具名称"
            value={newTool.name}
            onChange={(e) => setNewTool({ ...newTool, name: e.target.value })}
          />
          <input
            type="text"
            placeholder="描述"
            value={newTool.description}
            onChange={(e) => setNewTool({ ...newTool, description: e.target.value })}
          />
          <button type="button" onClick={addTool} className="btn-add">
            添加
          </button>
        </div>
        <div className="tag-list">
          {formData.tools?.map((tool, index) => (
            <span key={index} className="tag-item">
              {tool.name}
              <button type="button" onClick={() => removeTool(index)}>
                ×
              </button>
            </span>
          ))}
        </div>
      </div>

      <div className="form-section">
        <h3>权限配置</h3>
        <div className="inline-form">
          <input
            type="text"
            placeholder="资源"
            value={newPermission.resource}
            onChange={(e) => setNewPermission({ ...newPermission, resource: e.target.value })}
          />
          <select
            value={newPermission.action}
            onChange={(e) => setNewPermission({ ...newPermission, action: e.target.value })}
          >
            <option value="">选择操作</option>
            <option value="read">读取</option>
            <option value="write">写入</option>
            <option value="execute">执行</option>
            <option value="delete">删除</option>
          </select>
          <button type="button" onClick={addPermission} className="btn-add">
            添加
          </button>
        </div>
        <div className="tag-list">
          {formData.permissions?.map((perm, index) => (
            <span key={index} className="tag-item permission">
              {perm.resource}:{perm.action}
              <button type="button" onClick={() => removePermission(index)}>
                ×
              </button>
            </span>
          ))}
        </div>
      </div>

      <div className="form-section">
        <h3>标签</h3>
        <div className="inline-form">
          <input
            type="text"
            placeholder="标签"
            value={newTag}
            onChange={(e) => setNewTag(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), addTag())}
          />
          <button type="button" onClick={addTag} className="btn-add">
            添加
          </button>
        </div>
        <div className="tag-list">
          {formData.tags?.map((tag) => (
            <span key={tag} className="tag-item">
              {tag}
              <button type="button" onClick={() => removeTag(tag)}>
                ×
              </button>
            </span>
          ))}
        </div>
      </div>

      <div className="form-actions">
        <button type="button" onClick={onCancel} className="btn-secondary">
          取消
        </button>
        <button type="submit" disabled={loading} className="btn-primary">
          {loading ? '保存中...' : agent ? '更新' : '创建'}
        </button>
      </div>
    </form>
  );
};

export default AgentForm;