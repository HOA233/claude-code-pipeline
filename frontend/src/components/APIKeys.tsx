import React, { useState, useCallback } from 'react';
import { useToast } from './Toast';
import './APIKeys.css';

interface APIKey {
  id: string;
  name: string;
  key: string;
  permissions: string[];
  created_at: string;
  last_used?: string;
  expires_at?: string;
  is_admin: boolean;
}

interface APIKeysProps {
  onClose?: () => void;
}

export const APIKeys: React.FC<APIKeysProps> = ({ onClose }) => {
  const [keys, setKeys] = useState<APIKey[]>([
    {
      id: '1',
      name: 'Default Admin Key',
      key: 'sk-admin-xxxxxxxxxxxx',
      permissions: ['read', 'write', 'admin'],
      created_at: new Date().toISOString(),
      is_admin: true,
    },
  ]);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [newKey, setNewKey] = useState({
    name: '',
    permissions: [] as string[],
    expires_days: 0,
  });
  const [generatedKey, setGeneratedKey] = useState<string | null>(null);
  const { addToast } = useToast();

  const availablePermissions = [
    { id: 'read', label: '读取' },
    { id: 'write', label: '写入' },
    { id: 'execute', label: '执行' },
    { id: 'admin', label: '管理员' },
  ];

  const handleCreateKey = () => {
    if (!newKey.name.trim()) {
      addToast('请输入密钥名称', 'warning');
      return;
    }

    const generated = `sk-${Math.random().toString(36).substring(2, 15)}${Math.random().toString(36).substring(2, 15)}`;
    const key: APIKey = {
      id: Date.now().toString(),
      name: newKey.name,
      key: generated,
      permissions: newKey.permissions,
      created_at: new Date().toISOString(),
      expires_at: newKey.expires_days > 0
        ? new Date(Date.now() + newKey.expires_days * 86400000).toISOString()
        : undefined,
      is_admin: newKey.permissions.includes('admin'),
    };

    setKeys([...keys, key]);
    setGeneratedKey(generated);
    setNewKey({ name: '', permissions: [], expires_days: 0 });
    addToast('API 密钥已创建', 'success');
  };

  const handleDeleteKey = (id: string) => {
    if (!confirm('确定要删除此 API 密钥吗？')) return;
    setKeys(keys.filter((k) => k.id !== id));
    addToast('API 密钥已删除', 'success');
  };

  const handleCopyKey = (key: string) => {
    navigator.clipboard.writeText(key);
    addToast('已复制到剪贴板', 'success');
  };

  const handlePermissionToggle = (perm: string) => {
    setNewKey((prev) => ({
      ...prev,
      permissions: prev.permissions.includes(perm)
        ? prev.permissions.filter((p) => p !== perm)
        : [...prev.permissions, perm],
    }));
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('zh-CN');
  };

  const maskKey = (key: string) => {
    return key.substring(0, 10) + '...' + key.substring(key.length - 4);
  };

  return (
    <div className="api-keys">
      <div className="keys-header">
        <h3>API 密钥管理</h3>
        <button className="btn-create" onClick={() => setShowCreateForm(true)}>
          + 创建密钥
        </button>
      </div>

      {generatedKey && (
        <div className="generated-key-banner">
          <div className="banner-content">
            <strong>新密钥已生成！</strong>
            <p>请立即复制保存，此密钥只会显示一次</p>
            <code className="generated-key-value">{generatedKey}</code>
          </div>
          <div className="banner-actions">
            <button onClick={() => handleCopyKey(generatedKey)}>
              复制密钥
            </button>
            <button onClick={() => setGeneratedKey(null)}>
              我已保存
            </button>
          </div>
        </div>
      )}

      <div className="keys-list">
        {keys.map((key) => (
          <div key={key.id} className={`key-item ${key.is_admin ? 'admin' : ''}`}>
            <div className="key-info">
              <div className="key-name">
                {key.name}
                {key.is_admin && <span className="admin-badge">Admin</span>}
              </div>
              <div className="key-value">
                <code>{maskKey(key.key)}</code>
                <button
                  className="btn-copy"
                  onClick={() => handleCopyKey(key.key)}
                  title="复制密钥"
                >
                  📋
                </button>
              </div>
              <div className="key-permissions">
                {key.permissions.map((perm) => (
                  <span key={perm} className="permission-tag">
                    {availablePermissions.find((p) => p.id === perm)?.label || perm}
                  </span>
                ))}
              </div>
              <div className="key-meta">
                <span>创建: {formatDate(key.created_at)}</span>
                {key.last_used && <span>最后使用: {formatDate(key.last_used)}</span>}
                {key.expires_at && <span>过期: {formatDate(key.expires_at)}</span>}
              </div>
            </div>
            <div className="key-actions">
              <button
                className="btn-delete"
                onClick={() => handleDeleteKey(key.id)}
              >
                删除
              </button>
            </div>
          </div>
        ))}
      </div>

      {showCreateForm && (
        <div className="modal-overlay" onClick={() => setShowCreateForm(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>创建 API 密钥</h3>
              <button className="modal-close" onClick={() => setShowCreateForm(false)}>
                ×
              </button>
            </div>
            <div className="modal-body">
              <div className="form-group">
                <label>密钥名称 *</label>
                <input
                  type="text"
                  value={newKey.name}
                  onChange={(e) => setNewKey({ ...newKey, name: e.target.value })}
                  placeholder="例如: Production API Key"
                />
              </div>

              <div className="form-group">
                <label>权限</label>
                <div className="permissions-grid">
                  {availablePermissions.map((perm) => (
                    <label key={perm.id} className="permission-checkbox">
                      <input
                        type="checkbox"
                        checked={newKey.permissions.includes(perm.id)}
                        onChange={() => handlePermissionToggle(perm.id)}
                      />
                      <span>{perm.label}</span>
                    </label>
                  ))}
                </div>
              </div>

              <div className="form-group">
                <label>过期时间</label>
                <select
                  value={newKey.expires_days}
                  onChange={(e) => setNewKey({ ...newKey, expires_days: parseInt(e.target.value) })}
                >
                  <option value="0">永不过期</option>
                  <option value="7">7 天</option>
                  <option value="30">30 天</option>
                  <option value="90">90 天</option>
                  <option value="365">1 年</option>
                </select>
              </div>
            </div>
            <div className="modal-footer">
              <button className="btn-cancel" onClick={() => setShowCreateForm(false)}>
                取消
              </button>
              <button className="btn-primary" onClick={handleCreateKey}>
                创建密钥
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default APIKeys;