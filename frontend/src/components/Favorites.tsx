import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../api/client';
import type { Agent, Workflow } from '../types';
import './Favorites.css';

interface Favorite {
  id: string;
  type: 'agent' | 'workflow';
  target_id: string;
  target_name: string;
  added_at: string;
}

interface FavoritesProps {
  onExecute?: (type: 'agent' | 'workflow', id: string) => void;
}

export const Favorites: React.FC<FavoritesProps> = ({ onExecute }) => {
  const [favorites, setFavorites] = useState<Favorite[]>([]);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  // Load favorites from localStorage
  const loadFavorites = useCallback(() => {
    setLoading(true);
    try {
      const stored = localStorage.getItem('claude-platform-favorites');
      if (stored) {
        setFavorites(JSON.parse(stored));
      }
    } catch (error) {
      console.error('Failed to load favorites:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadFavorites();
  }, [loadFavorites]);

  const saveFavorites = (favs: Favorite[]) => {
    localStorage.setItem('claude-platform-favorites', JSON.stringify(favs));
    setFavorites(favs);
  };

  const addFavorite = (type: 'agent' | 'workflow', target: Agent | Workflow) => {
    const exists = favorites.find(
      (f) => f.type === type && f.target_id === target.id
    );
    if (exists) return;

    const newFavorite: Favorite = {
      id: `fav-${Date.now()}`,
      type,
      target_id: target.id,
      target_name: target.name,
      added_at: new Date().toISOString(),
    };

    saveFavorites([...favorites, newFavorite]);
  };

  const removeFavorite = (id: string) => {
    saveFavorites(favorites.filter((f) => f.id !== id));
  };

  const handleNavigate = (type: 'agent' | 'workflow', target_id: string) => {
    if (type === 'agent') {
      navigate(`/agents?id=${target_id}`);
    } else {
      navigate(`/workflows?id=${target_id}`);
    }
  };

  const handleExecute = (type: 'agent' | 'workflow', target_id: string) => {
    onExecute?.(type, target_id);
  };

  const getIcon = (type: 'agent' | 'workflow') => {
    return type === 'agent' ? '🤖' : '🔄';
  };

  if (loading) {
    return <div className="favorites-loading">加载中...</div>;
  }

  if (favorites.length === 0) {
    return (
      <div className="favorites-empty">
        <span className="empty-icon">⭐</span>
        <span>暂无收藏项目</span>
        <span className="hint">点击 Agent 或工作流详情页的收藏按钮添加</span>
      </div>
    );
  }

  return (
    <div className="favorites">
      <div className="favorites-header">
        <h3>收藏夹</h3>
        <span className="count">{favorites.length}</span>
      </div>
      <div className="favorites-list">
        {favorites.map((fav) => (
          <div key={fav.id} className="favorite-item">
            <div className="fav-icon">{getIcon(fav.type)}</div>
            <div className="fav-info">
              <span className="fav-name">{fav.target_name}</span>
              <span className="fav-type">{fav.type === 'agent' ? 'Agent' : '工作流'}</span>
            </div>
            <div className="fav-actions">
              <button
                className="action-btn execute"
                onClick={() => handleExecute(fav.type, fav.target_id)}
                title="执行"
              >
                ▶
              </button>
              <button
                className="action-btn navigate"
                onClick={() => handleNavigate(fav.type, fav.target_id)}
                title="查看"
              >
                →
              </button>
              <button
                className="action-btn remove"
                onClick={() => removeFavorite(fav.id)}
                title="移除收藏"
              >
                ×
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

// Hook for managing favorites
export const useFavorites = () => {
  const [favorites, setFavorites] = useState<Favorite[]>([]);

  useEffect(() => {
    const stored = localStorage.getItem('claude-platform-favorites');
    if (stored) {
      setFavorites(JSON.parse(stored));
    }
  }, []);

  const saveFavorites = (favs: Favorite[]) => {
    localStorage.setItem('claude-platform-favorites', JSON.stringify(favs));
    setFavorites(favs);
  };

  const isFavorite = (type: 'agent' | 'workflow', targetId: string) => {
    return favorites.some((f) => f.type === type && f.target_id === targetId);
  };

  const toggleFavorite = (type: 'agent' | 'workflow', target: Agent | Workflow) => {
    const exists = favorites.find(
      (f) => f.type === type && f.target_id === target.id
    );

    if (exists) {
      saveFavorites(favorites.filter((f) => f.id !== exists.id));
      return false;
    } else {
      const newFavorite: Favorite = {
        id: `fav-${Date.now()}`,
        type,
        target_id: target.id,
        target_name: target.name,
        added_at: new Date().toISOString(),
      };
      saveFavorites([...favorites, newFavorite]);
      return true;
    }
  };

  return { favorites, isFavorite, toggleFavorite };
};

export default Favorites;