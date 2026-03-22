import React, { useState, useEffect, useCallback } from 'react';
import api from '../api/client';
import { useToast } from '../components/Toast';
import './Webhooks.css';

interface Webhook {
  id: string;
  name: string;
  url: string;
  events: string[];
  enabled: boolean;
  secret?: string;
  created_at: string;
  last_triggered?: string;
  delivery_count: number;
}

interface WebhookDelivery {
  id: string;
  webhook_id: string;
  event: string;
  status: number;
  duration: number;
  timestamp: string;
}

function Webhooks() {
  const [webhooks, setWebhooks] = useState<Webhook[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [selectedWebhook, setSelectedWebhook] = useState<Webhook | null>(null);
  const [deliveries, setDeliveries] = useState<WebhookDelivery[]>([]);
  const [formData, setFormData] = useState({
    name: '',
    url: '',
    events: [] as string[],
    secret: '',
  });
  const { addToast } = useToast();

  const availableEvents = [
    { id: 'execution.started', label: '执行开始' },
    { id: 'execution.completed', label: '执行完成' },
    { id: 'execution.failed', label: '执行失败' },
    { id: 'workflow.created', label: '工作流创建' },
    { id: 'workflow.updated', label: '工作流更新' },
    { id: 'agent.created', label: 'Agent 创建' },
    { id: 'schedule.triggered', label: '定时任务触发' },
  ];

  const fetchWebhooks = useCallback(async () => {
    setLoading(true);
    try {
      const result = await api.listWebhooks({});
      setWebhooks(result.webhooks || []);
    } catch (error) {
      console.error('Failed to fetch webhooks:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchWebhooks();
  }, [fetchWebhooks]);

  const handleCreate = async () => {
    if (!formData.name || !formData.url || formData.events.length === 0) {
      addToast('请填写必填字段', 'warning');
      return;
    }

    try {
      await api.createWebhook({
        name: formData.name,
        url: formData.url,
        events: formData.events,
        secret: formData.secret || undefined,
      });
      addToast('Webhook 已创建', 'success');
      setShowCreateForm(false);
      setFormData({ name: '', url: '', events: [], secret: '' });
      fetchWebhooks();
    } catch (error) {
      console.error('Failed to create webhook:', error);
      addToast('创建失败', 'error');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('确定要删除此 Webhook 吗？')) return;
    try {
      await api.deleteWebhook(id);
      addToast('Webhook 已删除', 'success');
      fetchWebhooks();
    } catch (error) {
      console.error('Failed to delete webhook:', error);
      addToast('删除失败', 'error');
    }
  };

  const handleToggle = async (webhook: Webhook) => {
    try {
      await api.updateWebhook(webhook.id, { enabled: !webhook.enabled });
      addToast(`Webhook 已${webhook.enabled ? '禁用' : '启用'}`, 'success');
      fetchWebhooks();
    } catch (error) {
      console.error('Failed to toggle webhook:', error);
      addToast('操作失败', 'error');
    }
  };

  const handleViewDeliveries = async (webhook: Webhook) => {
    setSelectedWebhook(webhook);
    try {
      const result = await api.getWebhookDeliveries(webhook.id, 20);
      setDeliveries(result.deliveries || []);
    } catch (error) {
      console.error('Failed to fetch deliveries:', error);
      setDeliveries([]);
    }
  };

  const handleEventToggle = (eventId: string) => {
    setFormData((prev) => ({
      ...prev,
      events: prev.events.includes(eventId)
        ? prev.events.filter((e) => e !== eventId)
        : [...prev.events, eventId],
    }));
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('zh-CN');
  };

  return (
    <div className="webhooks-page">
      <div className="page-header">
        <h1>Webhook 管理</h1>
        <p>配置外部系统的回调通知</p>
      </div>

      <div className="webhooks-content">
        <div className="webhooks-list-section">
          <div className="section-header">
            <h3>Webhooks</h3>
            <button className="btn-create" onClick={() => setShowCreateForm(true)}>
              + 新建 Webhook
            </button>
          </div>

          {loading ? (
            <div className="loading">加载中...</div>
          ) : webhooks.length > 0 ? (
            <div className="webhooks-list">
              {webhooks.map((webhook) => (
                <div key={webhook.id} className={`webhook-item ${!webhook.enabled ? 'disabled' : ''}`}>
                  <div className="webhook-info">
                    <div className="webhook-header">
                      <span className="webhook-name">{webhook.name}</span>
                      <span className={`webhook-status ${webhook.enabled ? 'enabled' : 'disabled'}`}>
                        {webhook.enabled ? '启用' : '禁用'}
                      </span>
                    </div>
                    <div className="webhook-url">{webhook.url}</div>
                    <div className="webhook-events">
                      {webhook.events?.slice(0, 3).map((event) => (
                        <span key={event} className="event-tag">
                          {availableEvents.find((e) => e.id === event)?.label || event}
                        </span>
                      ))}
                      {webhook.events?.length > 3 && (
                        <span className="event-tag more">+{webhook.events.length - 3}</span>
                      )}
                    </div>
                    <div className="webhook-meta">
                      <span>触发次数: {webhook.delivery_count || 0}</span>
                      {webhook.last_triggered && (
                        <span>最后触发: {formatDate(webhook.last_triggered)}</span>
                      )}
                    </div>
                  </div>
                  <div className="webhook-actions">
                    <button className="btn-view" onClick={() => handleViewDeliveries(webhook)}>
                      查看记录
                    </button>
                    <button onClick={() => handleToggle(webhook)}>
                      {webhook.enabled ? '禁用' : '启用'}
                    </button>
                    <button className="btn-delete" onClick={() => handleDelete(webhook.id)}>
                      删除
                    </button>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="empty">
              <p>暂无 Webhook</p>
              <p className="hint">创建 Webhook 以接收事件通知</p>
            </div>
          )}
        </div>

        {/* Deliveries Panel */}
        {selectedWebhook && (
          <div className="deliveries-panel">
            <div className="panel-header">
              <h3>触发记录 - {selectedWebhook.name}</h3>
              <button className="btn-close" onClick={() => setSelectedWebhook(null)}>
                ×
              </button>
            </div>
            <div className="deliveries-list">
              {deliveries.length > 0 ? (
                deliveries.map((delivery) => (
                  <div key={delivery.id} className="delivery-item">
                    <div className="delivery-event">{delivery.event}</div>
                    <div className={`delivery-status status-${Math.floor(delivery.status / 100)}`}>
                      {delivery.status}
                    </div>
                    <div className="delivery-time">{formatDate(delivery.timestamp)}</div>
                  </div>
                ))
              ) : (
                <div className="empty-deliveries">暂无触发记录</div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Create Form */}
      {showCreateForm && (
        <div className="create-form-overlay" onClick={() => setShowCreateForm(false)}>
          <div className="create-form" onClick={(e) => e.stopPropagation()}>
            <h3>新建 Webhook</h3>

            <div className="form-group">
              <label>名称 *</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="输入 Webhook 名称"
              />
            </div>

            <div className="form-group">
              <label>URL *</label>
              <input
                type="url"
                value={formData.url}
                onChange={(e) => setFormData({ ...formData, url: e.target.value })}
                placeholder="https://example.com/webhook"
              />
            </div>

            <div className="form-group">
              <label>订阅事件 *</label>
              <div className="events-grid">
                {availableEvents.map((event) => (
                  <label key={event.id} className="event-checkbox">
                    <input
                      type="checkbox"
                      checked={formData.events.includes(event.id)}
                      onChange={() => handleEventToggle(event.id)}
                    />
                    <span>{event.label}</span>
                  </label>
                ))}
              </div>
            </div>

            <div className="form-group">
              <label>密钥 (可选)</label>
              <input
                type="text"
                value={formData.secret}
                onChange={(e) => setFormData({ ...formData, secret: e.target.value })}
                placeholder="用于签名验证"
              />
            </div>

            <div className="form-actions">
              <button className="btn-cancel" onClick={() => setShowCreateForm(false)}>
                取消
              </button>
              <button className="btn-submit" onClick={handleCreate}>
                创建
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default Webhooks;