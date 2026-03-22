import React, { useState, useEffect } from 'react';
import './ExportImport.css';

interface ExportData {
  version: string;
  exported_at: string;
  agents: any[];
  workflows: any[];
  schedules: any[];
  webhooks: any[];
}

interface ExportImportProps {
  onExport?: () => Promise<ExportData>;
  onImport?: (data: ExportData) => Promise<void>;
  type: 'agents' | 'workflows' | 'schedules' | 'webhooks' | 'all';
}

export const ExportImport: React.FC<ExportImportProps> = ({
  onExport,
  onImport,
  type,
}) => {
  const [exporting, setExporting] = useState(false);
  const [importing, setImporting] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [importPreview, setImportPreview] = useState<ExportData | null>(null);
  const [importError, setImportError] = useState<string | null>(null);

  const handleExport = async () => {
    setExporting(true);
    try {
      let data: ExportData;

      if (onExport) {
        data = await onExport();
      } else {
        // Default export from localStorage or mock
        data = {
          version: '1.0.0',
          exported_at: new Date().toISOString(),
          agents: [],
          workflows: [],
          schedules: [],
          webhooks: [],
        };
      }

      const blob = new Blob([JSON.stringify(data, null, 2)], {
        type: 'application/json',
      });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `claude-platform-${type}-export-${new Date().toISOString().split('T')[0]}.json`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Export failed:', error);
      alert('导出失败: ' + (error as Error).message);
    } finally {
      setExporting(false);
    }
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (event) => {
      try {
        const content = event.target?.result as string;
        const data = JSON.parse(content) as ExportData;

        // Validate structure
        if (!data.version || !data.exported_at) {
          throw new Error('Invalid export file format');
        }

        setImportPreview(data);
        setImportError(null);
      } catch (error) {
        setImportError('无效的导出文件格式: ' + (error as Error).message);
        setImportPreview(null);
      }
    };
    reader.readAsText(file);
  };

  const handleImport = async () => {
    if (!importPreview) return;

    setImporting(true);
    try {
      if (onImport) {
        await onImport(importPreview);
      }
      setShowImportModal(false);
      setImportPreview(null);
      alert('导入成功');
    } catch (error) {
      console.error('Import failed:', error);
      alert('导入失败: ' + (error as Error).message);
    } finally {
      setImporting(false);
    }
  };

  const getImportSummary = () => {
    if (!importPreview) return null;

    const counts = {
      agents: importPreview.agents?.length || 0,
      workflows: importPreview.workflows?.length || 0,
      schedules: importPreview.schedules?.length || 0,
      webhooks: importPreview.webhooks?.length || 0,
    };

    return counts;
  };

  return (
    <div className="export-import">
      <div className="ei-buttons">
        <button
          className="ei-btn export"
          onClick={handleExport}
          disabled={exporting}
        >
          <span className="icon">📤</span>
          <span>{exporting ? '导出中...' : '导出'}</span>
        </button>
        <button
          className="ei-btn import"
          onClick={() => setShowImportModal(true)}
          disabled={importing}
        >
          <span className="icon">📥</span>
          <span>导入</span>
        </button>
      </div>

      {showImportModal && (
        <div className="ei-modal-overlay" onClick={() => setShowImportModal(false)}>
          <div className="ei-modal" onClick={(e) => e.stopPropagation()}>
            <div className="ei-modal-header">
              <h3>导入数据</h3>
              <button className="close-btn" onClick={() => setShowImportModal(false)}>
                ×
              </button>
            </div>

            <div className="ei-modal-body">
              {!importPreview ? (
                <div className="import-upload">
                  <label className="upload-area">
                    <input
                      type="file"
                      accept=".json"
                      onChange={handleFileSelect}
                    />
                    <div className="upload-content">
                      <span className="upload-icon">📁</span>
                      <span>点击选择或拖拽文件到此处</span>
                      <span className="hint">支持 .json 格式的导出文件</span>
                    </div>
                  </label>
                  {importError && (
                    <div className="import-error">
                      <span>❌</span>
                      <span>{importError}</span>
                    </div>
                  )}
                </div>
              ) : (
                <div className="import-preview">
                  <div className="preview-header">
                    <span className="preview-icon">✅</span>
                    <span>文件验证通过</span>
                  </div>

                  <div className="preview-meta">
                    <div className="meta-item">
                      <span className="label">版本</span>
                      <span className="value">{importPreview.version}</span>
                    </div>
                    <div className="meta-item">
                      <span className="label">导出时间</span>
                      <span className="value">
                        {new Date(importPreview.exported_at).toLocaleString('zh-CN')}
                      </span>
                    </div>
                  </div>

                  <div className="preview-counts">
                    <h4>将导入以下内容</h4>
                    <div className="counts-grid">
                      {getImportSummary()?.agents ? (
                        <div className="count-item">
                          <span className="count-icon">🤖</span>
                          <span className="count-label">Agents</span>
                          <span className="count-value">{getImportSummary()?.agents}</span>
                        </div>
                      ) : null}
                      {getImportSummary()?.workflows ? (
                        <div className="count-item">
                          <span className="count-icon">🔄</span>
                          <span className="count-label">Workflows</span>
                          <span className="count-value">{getImportSummary()?.workflows}</span>
                        </div>
                      ) : null}
                      {getImportSummary()?.schedules ? (
                        <div className="count-item">
                          <span className="count-icon">⏰</span>
                          <span className="count-label">Schedules</span>
                          <span className="count-value">{getImportSummary()?.schedules}</span>
                        </div>
                      ) : null}
                      {getImportSummary()?.webhooks ? (
                        <div className="count-item">
                          <span className="count-icon">🔔</span>
                          <span className="count-label">Webhooks</span>
                          <span className="count-value">{getImportSummary()?.webhooks}</span>
                        </div>
                      ) : null}
                    </div>
                  </div>

                  <div className="import-warning">
                    <span className="warning-icon">⚠️</span>
                    <span>导入将覆盖同名的现有配置</span>
                  </div>
                </div>
              )}
            </div>

            {importPreview && (
              <div className="ei-modal-footer">
                <button
                  className="btn-cancel"
                  onClick={() => {
                    setImportPreview(null);
                    setImportError(null);
                  }}
                >
                  取消
                </button>
                <button
                  className="btn-import"
                  onClick={handleImport}
                  disabled={importing}
                >
                  {importing ? '导入中...' : '确认导入'}
                </button>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default ExportImport;