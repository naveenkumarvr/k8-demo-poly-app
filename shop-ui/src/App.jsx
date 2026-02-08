import React from 'react';
import { Routes, Route, Link } from 'react-router-dom';
import { ShoppingCart, Package } from 'lucide-react';
import Home from './pages/Home';
import CartOverlay from './components/CartOverlay';
import { CartProvider, useCart } from './context/CartContext';

function Layout({ children }) {
  const { cartItems, setIsOpen } = useCart();
  const itemCount = cartItems.reduce((acc, item) => acc + item.quantity, 0);

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col font-sans text-gray-900">
      <header className="bg-white shadow-sm sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2 text-blue-600 hover:text-blue-700 transition-colors">
            <Package className="w-8 h-8" />
            <span className="font-bold text-xl tracking-tight">PolyShop</span>
          </Link>

          <button
            onClick={() => setIsOpen(true)}
            className="relative p-2 hover:bg-gray-100 rounded-full transition-colors"
            aria-label="Open cart"
          >
            <ShoppingCart className="w-6 h-6 text-gray-700" />
            {itemCount > 0 && (
              <span className="absolute top-0 right-0 bg-red-500 text-white text-xs font-bold w-5 h-5 flex items-center justify-center rounded-full ring-2 ring-white">
                {itemCount}
              </span>
            )}
          </button>
        </div>
      </header>

      <main className="flex-grow max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {children}
      </main>

      <footer className="bg-white border-t py-8 mt-auto">
        <div className="max-w-7xl mx-auto px-4 text-center text-gray-500 text-sm">
          &copy; {new Date().getFullYear()} PolyShop Microservices Demo. Built with React, Vite & Tailwind.
        </div>
      </footer>

      <CartOverlay />
    </div>
  );
}

function App() {
  return (
    <CartProvider>
      <Layout>
        <Routes>
          <Route path="/" element={<Home />} />
        </Routes>
      </Layout>
    </CartProvider>
  );
}

export default App;
