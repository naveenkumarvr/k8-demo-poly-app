import React, { createContext, useContext, useState, useEffect } from 'react';

const CartContext = createContext();

export const useCart = () => {
    const context = useContext(CartContext);
    if (!context) {
        throw new Error('useCart must be used within a CartProvider');
    }
    return context;
};

export const CartProvider = ({ children }) => {
    const [cartItems, setCartItems] = useState([]);
    const [isOpen, setIsOpen] = useState(false);
    const [isLoading, setIsLoading] = useState(false);

    const fetchCart = async () => {
        setIsLoading(true);
        try {
            const response = await fetch('/api/cart'); // Proxy to cart-service
            if (response.ok) {
                const data = await response.json();
                // Assuming cart usage is simplified: array of items or object with items
                setCartItems(Array.isArray(data) ? data : (data.items || []));
            }
        } catch (error) {
            console.error('Failed to fetch cart:', error);
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchCart();
    }, []);

    const addToCart = async (product) => {
        try {
            await fetch('/api/cart', { // POST /api/cart for adding item? Check service spec later but assume standard REST
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ item: { product_id: product.id, quantity: 1 } }), // Structure depends on API
            });
            await fetchCart();
            setIsOpen(true);
        } catch (error) {
            console.error('Failed to add to cart:', error);
        }
    };

    const checkout = async () => {
        try {
            // Implement checkout logic calling checkout-service
            const response = await fetch('/api/checkout', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    email: 'user@example.com', // Mock user
                    address: {}
                }),
            });
            if (response.ok) {
                setCartItems([]);
                setIsOpen(false);
                return true;
            }
        } catch (e) {
            console.error("Checkout error", e);
        }
        return false;
    };

    return (
        <CartContext.Provider value={{ cartItems, addToCart, checkout, isOpen, setIsOpen, fetchCart, isLoading }}>
            {children}
        </CartContext.Provider>
    );
};
