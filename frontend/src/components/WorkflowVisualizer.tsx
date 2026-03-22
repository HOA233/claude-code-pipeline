import React, { useMemo } from 'react';
import type { Workflow, AgentNode, Connection } from '../types';
import './WorkflowVisualizer.css';

interface WorkflowVisualizerProps {
  workflow: Workflow;
  executionStatus?: Record<string, string>;
  highlightNode?: string;
  onNodeClick?: (node: AgentNode) => void;
}

export const WorkflowVisualizer: React.FC<WorkflowVisualizerProps> = ({
  workflow,
  executionStatus,
  highlightNode,
  onNodeClick,
}) => {
  const { nodes, edges, levels } = useMemo(() => {
    const nodeList: Array<AgentNode & { level: number; x: number; y: number }> = [];
    const edgeList: { from: string; to: string; fromX: number; fromY: number; toX: number; toY: number }[] = [];

    // Build dependency graph
    const nodeMap = new Map<string, AgentNode>();
    const dependencies = new Map<string, string[]>();
    const dependents = new Map<string, string[]>();

    workflow.agents.forEach((agent) => {
      nodeMap.set(agent.id, agent);
      dependencies.set(agent.id, agent.depends_on || []);
      dependents.set(agent.id, []);
    });

    // Build reverse dependencies
    workflow.agents.forEach((agent) => {
      (agent.depends_on || []).forEach((dep) => {
        if (dependents.has(dep)) {
          dependents.get(dep)!.push(agent.id);
        }
      });
    });

    // Calculate levels (topological sort)
    const levels = new Map<string, number>();
    const queue: string[] = [];

    // Find root nodes (no dependencies)
    workflow.agents.forEach((agent) => {
      if (!agent.depends_on || agent.depends_on.length === 0) {
        levels.set(agent.id, 0);
        queue.push(agent.id);
      }
    });

    // BFS to assign levels
    while (queue.length > 0) {
      const current = queue.shift()!;
      const currentLevel = levels.get(current)!;

      dependents.get(current)?.forEach((dep) => {
        const existingLevel = levels.get(dep) || 0;
        const newLevel = Math.max(existingLevel, currentLevel + 1);
        levels.set(dep, newLevel);

        if (!queue.includes(dep)) {
          queue.push(dep);
        }
      });
    }

    // Group nodes by level
    const levelGroups = new Map<number, AgentNode[]>();
    workflow.agents.forEach((agent) => {
      const level = levels.get(agent.id) || 0;
      if (!levelGroups.has(level)) {
        levelGroups.set(level, []);
      }
      levelGroups.get(level)!.push(agent);
    });

    // Calculate positions
    const nodeWidth = 200;
    const nodeHeight = 80;
    const horizontalGap = 60;
    const verticalGap = 100;

    let maxY = 0;
    levelGroups.forEach((nodesAtLevel, level) => {
      const totalWidth = nodesAtLevel.length * nodeWidth + (nodesAtLevel.length - 1) * horizontalGap;
      const startX = -totalWidth / 2;
      const y = level * (nodeHeight + verticalGap);

      nodesAtLevel.forEach((agent, index) => {
        const x = startX + index * (nodeWidth + horizontalGap);
        nodeList.push({
          ...agent,
          level,
          x: x + nodeWidth / 2,
          y: y + nodeHeight / 2,
        });
        maxY = Math.max(maxY, y + nodeHeight);
      });
    });

    // Create edges
    workflow.agents.forEach((agent) => {
      (agent.depends_on || []).forEach((dep) => {
        const fromNode = nodeList.find((n) => n.id === dep);
        const toNode = nodeList.find((n) => n.id === agent.id);
        if (fromNode && toNode) {
          edgeList.push({
            from: dep,
            to: agent.id,
            fromX: fromNode.x,
            fromY: fromNode.y + nodeHeight / 2,
            toX: toNode.x,
            toY: toNode.y - nodeHeight / 2,
          });
        }
      });
    });

    return { nodes: nodeList, edges: edgeList, levels: levelGroups };
  }, [workflow]);

  const getStatusColor = (nodeId: string) => {
    if (!executionStatus) return undefined;
    const status = executionStatus[nodeId];
    switch (status) {
      case 'completed':
        return '#52c41a';
      case 'running':
        return '#1890ff';
      case 'failed':
        return '#f5222d';
      case 'pending':
        return '#8c8c8c';
      default:
        return undefined;
    }
  };

  const getModeIcon = () => {
    switch (workflow.mode) {
      case 'serial':
        return '→';
      case 'parallel':
        return '⇉';
      case 'hybrid':
        return '⇄';
      default:
        return '○';
    }
  };

  // Calculate viewBox
  const minX = Math.min(...nodes.map((n) => n.x)) - 150;
  const maxX = Math.max(...nodes.map((n) => n.x)) + 150;
  const minY = -50;
  const maxY = Math.max(...nodes.map((n) => n.y)) + 100;
  const width = maxX - minX;
  const height = maxY - minY;

  return (
    <div className="workflow-visualizer">
      <div className="visualizer-header">
        <h4>{workflow.name}</h4>
        <span className="mode-badge">
          <span className="mode-icon">{getModeIcon()}</span>
          {workflow.mode}
        </span>
      </div>

      <div className="visualizer-container">
        <svg
          viewBox={`${minX} ${minY} ${width} ${height}`}
          className="workflow-svg"
        >
          {/* Edges (arrows) */}
          <defs>
            <marker
              id="arrowhead"
              markerWidth="10"
              markerHeight="7"
              refX="9"
              refY="3.5"
              orient="auto"
            >
              <polygon points="0 0, 10 3.5, 0 7" fill="#78716C" />
            </marker>
            <marker
              id="arrowhead-active"
              markerWidth="10"
              markerHeight="7"
              refX="9"
              refY="3.5"
              orient="auto"
            >
              <polygon points="0 0, 10 3.5, 0 7" fill="#1890ff" />
            </marker>
          </defs>

          {edges.map((edge, index) => {
            const isActive =
              executionStatus?.[edge.from] === 'completed' &&
              (executionStatus?.[edge.to] === 'running' ||
                executionStatus?.[edge.to] === 'completed');
            return (
              <path
                key={index}
                d={`M ${edge.fromX} ${edge.fromY}
                    C ${edge.fromX} ${edge.fromY + 40},
                      ${edge.toX} ${edge.toY - 40},
                      ${edge.toX} ${edge.toY}`}
                stroke={isActive ? '#1890ff' : '#78716C'}
                strokeWidth={isActive ? 2 : 1}
                fill="none"
                markerEnd={isActive ? 'url(#arrowhead-active)' : 'url(#arrowhead)'}
              />
            );
          })}

          {/* Nodes */}
          {nodes.map((node) => {
            const statusColor = getStatusColor(node.id);
            const isHighlighted = highlightNode === node.id;

            return (
              <g
                key={node.id}
                transform={`translate(${node.x - 100}, ${node.y - 40})`}
                className={`workflow-node ${isHighlighted ? 'highlighted' : ''}`}
                onClick={() => onNodeClick?.(node)}
              >
                <rect
                  width="200"
                  height="80"
                  rx="8"
                  fill={statusColor ? `${statusColor}20` : '#1C1917'}
                  stroke={statusColor || (isHighlighted ? '#1890ff' : '#44403C')}
                  strokeWidth={isHighlighted ? 2 : 1}
                />
                <text
                  x="100"
                  y="30"
                  textAnchor="middle"
                  fill="#FAFAF9"
                  fontSize="14"
                  fontWeight="500"
                >
                  {node.name || node.id}
                </text>
                <text
                  x="100"
                  y="50"
                  textAnchor="middle"
                  fill="#78716C"
                  fontSize="12"
                >
                  {node.agent_id}
                </text>
                {executionStatus?.[node.id] && (
                  <text
                    x="100"
                    y="68"
                    textAnchor="middle"
                    fill={statusColor}
                    fontSize="10"
                  >
                    {executionStatus[node.id]}
                  </text>
                )}
              </g>
            );
          })}
        </svg>
      </div>

      <div className="visualizer-legend">
        <span className="legend-item">
          <span className="legend-dot completed" /> Completed
        </span>
        <span className="legend-item">
          <span className="legend-dot running" /> Running
        </span>
        <span className="legend-item">
          <span className="legend-dot failed" /> Failed
        </span>
        <span className="legend-item">
          <span className="legend-dot pending" /> Pending
        </span>
      </div>
    </div>
  );
};

export default WorkflowVisualizer;