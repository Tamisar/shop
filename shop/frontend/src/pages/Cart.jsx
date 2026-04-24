import { useCart } from '../CartContext';

export default function Cart() {
  const { items } = useCart();
  return (
    <div style={{ padding: '2rem' }}>
      <h2>Корзина</h2>
      {items.length === 0 ? <p>Корзина пуста</p> : (
        <ul>
          {items.map((item, i) => (
            <li key={i}>Товар ID: {item.product_id} | Кол-во: {item.quantity}</li>
          ))}
        </ul>
      )}
    </div>
  );
}