import React, { useState } from 'react';
import api from '../api/client';
import type { Workflow, Agent, WorkflowCreateRequest, AgentNode } from '../types';
import './WorkflowBuilder.css';

interface WorkflowBuilderProps {
  agents: Agent[];
  onSave: (workflow: Workflow) => void;
  onCancel: () => void;
  initialWorkflow?: Workflow;
}

interface DragNode {
  id: string;
  agent_id: string;
  x: number;
  y: number;
}

export const WorkflowBuilder: React.FC<WorkflowBuilderProps> = ({
  agents,
  onSave,
  onCancel,
  initialWorkflow,
}) => {
  const [name, setName] = useState(initialWorkflow?.name || '');
  const [description, setDescription] = useState(initialWorkflow?.description || '');
  const [mode, setMode] = useState<'serial' | 'parallel' | 'hybrid'>(
    (initialWorkflow?.mode as any) || 'serial'
  );
  const [nodes, setNodes] = useState<DragNode[]>([]);
  const [connections, setConnections] = useState<{ from: string; to: string }[]>([]);
  const [selectedNode, setSelectedNode] = useState<string | null>(null);
  const [draggedAgent, setDraggedAgent] = useState<string | null>(null);

  const handleDragStart = (agentId: string) => {
    setDraggedAgent(agentId);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    if (!draggedAgent) return;

    const rect = e.currentTarget.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;

    const newNode: DragNode = {
      id: `node-${Date.now()}`,
      agent_id: draggedAgent,
      x,
      y,
    };

    setNodes([...nodes, newNode]);
    setDraggedAgent(null);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
  };

  const handleNodeClick = (nodeId: string) => {
    if (selectedNode === null) {
      setSelectedNode(nodeId);
    } else if (selectedNode !== nodeId) {
      // Create connection
      setConnections([...connections, { from: selectedNode, to: nodeId }]);
      setSelectedNode(null);
    } else {
      setSelectedNode(null);
    }
  };

  const handleNodeDelete = (nodeId: string) => {
    setNodes(nodes.filter((n) => n.id !== nodeId));
    setConnections(
      connections.filter((c) => c.from !== nodeId && c.to !== nodeId)
    );
  };

  const handleSave = async () => {
    if (!name.trim()) {
      alert('请输入工作流名称');
      return;
    }

    const workflowNodes: AgentNode[] = nodes.map((node, index) => {
      const dependsOn = connections
        .filter((c) => c.to === node.id)
        .map((c) => c.from);

      return {
        id: node.id,
        agent_id: node.agent_id,
        name: agents.find((a) => a.id === node.agent_id)?.name || '',
        depends_on: dependsOn.length > 0 ? dependsOn : undefined,
      };
    });

    const request: WorkflowCreateRequest = {
      name,
      description,
      mode,
      agents: workflowNodes,
      connections: connections.map((c) => ({
        from_node: c.from,
        from_output: 'output',
        to_node: c.to,
        to_input: 'input',
      })),
    };

    try {
      const workflow = initialWorkflow
        ? await api.updateWorkflow(initialWorkflow.id, request)
        : await api.createWorkflow(request);
      onSave(workflow);
    } catch (error) {
      console.error('Failed to save workflow:', error);
      alert('保存失败');
    }
  };

  const clearCanvas = () => {
    setNodes([]);
    setConnections([]);
    setSelectedNode(null);
  };

  const handleExport = () => {
    const config = {
      name,
      description,
      mode,
      nodes,
      connections,
      exportedAt: new Date().toISOString(),
    };
    const blob = new Blob([JSON.stringify(config, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `workflow-${name || 'config'}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleImport = () => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.json';
    input.onchange = (e: any) => {
      const file = e.target.files[0];
      if (!file) return;

      const reader = new FileReader();
      reader.onload = (event) => {
        try {
          const config = JSON.parse(event.target?.result as string);
          if (config.name) setName(config.name);
          if (config.description) setDescription(config.description);
          if (config.mode) setMode(config.mode);
          if (config.nodes) setNodes(config.nodes);
          if (config.connections) setConnections(config.connections);
        } catch (err) {
          alert('导入失败：无效的配置文件');
        }
      };
      reader.readAsText(file);
    };
    input.click();
  };

  return (
    <div className="workflow-builder">
      <div className="builder-sidebar">
        <h3>Agent 列表</h3>
        <p className="hint">拖拽 Agent 到画布</p>
        <div className="agent-palette">
          {agents.map((agent) => (
            <div
              key={agent.id}
              className="palette-item"
              draggable
              onDragStart={() => handleDragStart(agent.id)}
            >
              <span className="agent-icon">🤖</span>
              <span className="agent-name">{agent.name}</span>
              <span className="agent-category">{agent.category}</span>
            </div>
          ))}
        </div>
      </div>

      <div className="builder-main">
        <div className="builder-toolbar">
          <div className="toolbar-left">
            <input
              type="text"
              placeholder="工作流名称"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="workflow-name-input"
            />
            <select value={mode} onChange={(e) => setMode(e.target.value as any)}>
              <option value="serial">串行执行</option>
              <option value="parallel">并行执行</option>
              <option value="hybrid">混合模式</option>
            </select>
          </div>
          <div className="toolbar-right">
            <button onClick={handleImport} className="btn-secondary">
              导入配置
            </button>
            <button onClick={handleExport} className="btn-secondary">
              导出配置
            </button>
            <button onClick={clearCanvas} className="btn-secondary">
              清空画布
            </button>
            <button onClick={onCancel} className="btn-secondary">
              取消
            </button>
            <button onClick={handleSave} className="btn-primary">
              保存工作流
            </button>
          </div>
        </div>

        <div
          className="builder-canvas"
          onDrop={handleDrop}
          onDragOver={handleDragOver}
        >
          <svg className="connections-layer">
            {connections.map((conn, index) => {
              const fromNode = nodes.find((n) => n.id === conn.from);
              const toNode = nodes.find((n) => n.id === conn.to);
              if (!fromNode || !toNode) return null;

              return (
                <line
                  key={index}
                  x1={fromNode.x + 75}
                  y1={fromNode.y + 40}
                  x2={toNode.x + 75}
                  y2={toNode.y + 40}
                  className="connection-line"
                  markerEnd="url(#arrowhead)"
                />
              );
            })}
            <defs>
              <marker
                id="arrowhead"
                markerWidth="10"
                markerHeight="7"
                refX="9"
                refY="3.5"
                orient="auto"
              >
                <polygon points="0 0, 10 3.5, 0 7" fill="#666" />
              </marker>
            </defs>
          </svg>

          {nodes.map((node) => {
            const agent = agents.find((a) => a.id === node.agent_id);
            const isSelected = selectedNode === node.id;
            const hasIncoming = connections.some((c) => c.to === node.id);
            const hasOutgoing = connections.some((c) => c.from === node.id);

            return (
              <div
                key={node.id}
                className={`canvas-node ${isSelected ? 'selected' : ''}`}
                style={{ left: node.x, top: node.y }}
                onClick={() => handleNodeClick(node.id)}
              >
                <div className="node-header">
                  <span className="node-icon">🤖</span>
                  <span className="node-title">{agent?.name || 'Unknown'}</span>
                  <button
                    className="node-delete"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleNodeDelete(node.id);
                    }}
                  >
                    ×
                  </button>
                </div>
                <div className="node-body">
                  <span className="node-category">{agent?.category}</span>
                  <div className="node-indicators">
                    {hasIncoming && <span className="indicator in">← 输入</span>}
                    {hasOutgoing && <span className="indicator out">输出 →</span>}
                  </div>
                </div>
              </div>
            );
          })}

          {nodes.length === 0 && (
            <div className="canvas-empty">
              <div className="empty-icon">📋</div>
              <p>拖拽 Agent 到此处开始构建工作流</p>
            </div>
          )}
        </div>

        <div className="builder-info">
          <textarea
            placeholder="工作流描述..."
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className="description-input"
          />
          <div className="stats">
            <span>节点: {nodes.length}</span>
            <span>连接: {connections.length}</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default WorkflowBuilder;