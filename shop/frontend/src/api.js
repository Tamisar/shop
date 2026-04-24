const API_BASE_URL = import.meta.env.VITE_API_URL ?? '';

export const api = fetch => (url, options = {}) => {
  const token = localStorage.getItem('token');
  const headers = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options.headers
  };
  return fetch(`${API_BASE_URL}${url}`, { ...options, headers });
};
