import { createContext, useContext, useState, useEffect } from 'react';
import { api } from './api';

const CartContext = createContext(null);

export const CartProvider = ({ children }) => {
  const [items, setItems] = useState([]);
  const [cartApi] = useState(() => api(fetch));

  useEffect(() => {
    if (localStorage.getItem('token')) {
      cartApi('/api/cart').then(res => res.json()).then(data => setItems(data.items || []));
    }
  }, []);

  const addToCart = async (productId) => {
    await cartApi('/api/cart', { method: 'POST', body: JSON.stringify({ product_id: productId, quantity: 1 }) });
    setItems(prev => [...prev, { product_id: productId, quantity: 1 }]);
  };

  return <CartContext.Provider value={{ items, addToCart }}>{children}</CartContext.Provider>;
};

export const useCart = () => useContext(CartContext);