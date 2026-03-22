import React, { useState, useEffect } from 'react';
import './NotificationBanner.css';

interface Banner {
  id: string;
  type: 'info' | 'warning' | 'error' | 'success';
  message: string;
  link?: string;
  linkText?: string;
  dismissible: boolean;
  expiresAt?: string;
}

interface NotificationBannerProps {
  position?: 'top' | 'bottom';
}

export const NotificationBanner: React.FC<NotificationBannerProps> = ({
  position = 'top',
}) => {
  const [banners, setBanners] = useState<Banner[]>([]);
  const [dismissed, setDismissed] = useState<Set<string>>(new Set());

  useEffect(() => {
    // Load dismissed banners from localStorage
    const stored = localStorage.getItem('claude-platform-dismissed-banners');
    if (stored) {
      setDismissed(new Set(JSON.parse(stored)));
    }

    // Load active banners (could be from API)
    // For now, we'll use sample banners
    const activeBanners: Banner[] = [
      {
        id: 'system-online',
        type: 'success',
        message: '系统运行正常',
        dismissible: true,
      },
    ];

    // Filter out expired and dismissed banners
    const now = new Date();
    const visibleBanners = activeBanners.filter((banner) => {
      if (dismissed.has(banner.id)) return false;
      if (banner.expiresAt && new Date(banner.expiresAt) < now) return false;
      return true;
    });

    setBanners(visibleBanners);
  }, []);

  const handleDismiss = (id: string) => {
    const newDismissed = new Set(dismissed);
    newDismissed.add(id);
    setDismissed(newDismissed);
    localStorage.setItem(
      'claude-platform-dismissed-banners',
      JSON.stringify(Array.from(newDismissed))
    );
    setBanners(banners.filter((b) => b.id !== id));
  };

  const getIcon = (type: Banner['type']) => {
    switch (type) {
      case 'info': return 'ℹ️';
      case 'warning': return '⚠️';
      case 'error': return '❌';
      case 'success': return '✅';
      default: return 'ℹ️';
    }
  };

  if (banners.length === 0) return null;

  return (
    <div className={`notification-banners ${position}`}>
      {banners.map((banner) => (
        <div key={banner.id} className={`banner ${banner.type}`}>
          <span className="banner-icon">{getIcon(banner.type)}</span>
          <span className="banner-message">{banner.message}</span>
          {banner.link && (
            <a href={banner.link} className="banner-link" target="_blank" rel="noopener noreferrer">
              {banner.linkText || '了解更多'}
            </a>
          )}
          {banner.dismissible && (
            <button
              className="banner-dismiss"
              onClick={() => handleDismiss(banner.id)}
              aria-label="关闭"
            >
              ×
            </button>
          )}
        </div>
      ))}
    </div>
  );
};

// Hook for programmatically showing banners
export const useBanner = () => {
  const [customBanners, setCustomBanners] = useState<Banner[]>([]);

  const showBanner = (banner: Omit<Banner, 'id'>) => {
    const newBanner: Banner = {
      ...banner,
      id: `custom-${Date.now()}`,
      dismissible: banner.dismissible ?? true,
    };
    setCustomBanners((prev) => [...prev, newBanner]);

    // Auto dismiss after 10 seconds for non-error types
    if (banner.type !== 'error') {
      setTimeout(() => {
        setCustomBanners((prev) => prev.filter((b) => b.id !== newBanner.id));
      }, 10000);
    }
  };

  const dismissBanner = (id: string) => {
    setCustomBanners((prev) => prev.filter((b) => b.id !== id));
  };

  return { customBanners, showBanner, dismissBanner };
};

export default NotificationBanner;