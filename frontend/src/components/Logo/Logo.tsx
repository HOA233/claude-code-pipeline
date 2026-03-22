import React from 'react'

const Logo = ({ size = 40, className = '' }) => {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 40 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
    >
      <defs>
        <linearGradient id="logoGradient" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="#D97706" />
          <stop offset="100%" stopColor="#FB923C" />
        </linearGradient>
      </defs>
      <rect width="40" height="40" rx="10" fill="url(#logoGradient)" />
      <path
        d="M12 20C12 15.5817 15.5817 12 20 12V12C24.4183 12 28 15.5817 28 20V28H20C15.5817 28 12 24.4183 12 20V20Z"
        fill="white"
        fillOpacity="0.9"
      />
    </svg>
  )
}

export default Logo