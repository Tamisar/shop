import { useState } from 'react';
import { useAuth } from '../AuthContext';
import { api } from '../api';
import { useNavigate } from 'react-router-dom';

export default function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const { login } = useAuth();
  const navigate = useNavigate();
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const res = await api(fetch)('/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password })
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error);
      login(email, data.token);
      navigate('/catalog');
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <form onSubmit={handleSubmit} style={{ maxWidth: 300, margin: '2rem auto' }}>
      <h2>Вход</h2>
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <input placeholder="Email" value={email} onChange={e => setEmail(e.target.value)} required style={{ display: 'block', margin: '0.5rem 0', width: '100%' }} />
      <input type="password" placeholder="Пароль" value={password} onChange={e => setPassword(e.target.value)} required style={{ display: 'block', margin: '0.5rem 0', width: '100%' }} />
      <button type="submit">Войти</button>
    </form>
  );
}