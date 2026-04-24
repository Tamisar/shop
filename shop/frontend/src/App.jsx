import { BrowserRouter, Routes, Route, Navigate, Link } from 'react-router-dom';
import { AuthProvider, useAuth } from './AuthContext';
import { CartProvider, useCart } from './CartContext';
import Login from './pages/Login';
import Register from './pages/Register';
import Catalog from './pages/Catalog';
import Cart from './pages/Cart';

function Header() {
  const { user, logout } = useAuth();
  return (
    <nav style={{ display: 'flex', justifyContent: 'space-between', padding: '1rem', background: '#f8f9fa', alignItems: 'center' }}>
      <div>
        <Link to="/catalog" style={{ marginRight: '1rem' }}>Каталог</Link>
        <Link to="/cart">Корзина ({useCart().items.length})</Link>
      </div>
      <div>
        {user ? (
          <>
            <span style={{ marginRight: '1rem' }}>{user.email}</span>
            <button onClick={logout}>Выйти</button>
          </>
        ) : <Link to="/login">Войти</Link>}
      </div>
    </nav>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <CartProvider>
          <Header />
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/catalog" element={<Catalog />} />
            <Route path="/cart" element={<Cart />} />
            <Route path="*" element={<Navigate to="/catalog" />} />
          </Routes>
        </CartProvider>
      </AuthProvider>
    </BrowserRouter>
  );
}
