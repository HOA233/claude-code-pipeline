import React from 'react';
import { useFavorites } from './Favorites';
import type { Agent, Workflow } from '../types';
import './FavoriteButton.css';

interface FavoriteButtonProps {
  type: 'agent' | 'workflow';
  target: Agent | Workflow;
  size?: 'small' | 'medium' | 'large';
  showLabel?: boolean;
}

export const FavoriteButton: React.FC<FavoriteButtonProps> = ({
  type,
  target,
  size = 'medium',
  showLabel = false,
}) => {
  const { isFavorite, toggleFavorite } = useFavorites();
  const isFav = isFavorite(type, target.id);

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    e.preventDefault();
    toggleFavorite(type, target);
  };

  return (
    <button
      className={`favorite-btn ${size} ${isFav ? 'active' : ''}`}
      onClick={handleClick}
      title={isFav ? '移除收藏' : '添加收藏'}
    >
      <span className="star-icon">{isFav ? '★' : '☆'}</span>
      {showLabel && (
        <span className="label">{isFav ? '已收藏' : '收藏'}</span>
      )}
    </button>
  );
};

export default FavoriteButton;