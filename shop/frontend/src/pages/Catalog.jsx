import { useEffect, useState } from 'react';
import { api } from '../api';
import { useCart } from '../CartContext';
import { useAuth } from '../AuthContext';

export default function Catalog() {
  const [products, setProducts] = useState([]);
  const { addToCart } = useCart();
  const { user } = useAuth();

  useEffect(() => {
    api(fetch)('/api/products').then(res => res.json()).then(setProducts);
  }, []);

  if (!user) return <p style={{ textAlign: 'center', marginTop: '2rem' }}>Войдите для просмотра каталога</p>;

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(250px, 1fr))', gap: '1rem', padding: '2rem' }}>
      {products.map(p => (
        <div key={p.id} style={{ border: '1px solid #ccc', borderRadius: 8, padding: '1rem', textAlign: 'center' }}>
          <img src={p.image} alt={p.name} style={{ width: '100%', height: 150, objectFit: 'cover', borderRadius: 4 }} />
          <h3>{p.name}</h3>
          <p>${p.price}</p>
          <button onClick={() => addToCart(p.id)} style={{ background: '#007bff', color: '#fff', border: 'none', padding: '0.5rem 1rem', borderRadius: 4, cursor: 'pointer' }}>В корзину</button>
        </div>
      ))}
    </div>
  );
}