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
    const [checkoutResponse, setCheckoutResponse] = useState(null);

    const fetchCart = async () => {
        setIsLoading(true);
        try {
            const response = await fetch('/api/cart/guest'); // GET /cart/:user_id
            if (response.ok) {
                const cartData = await response.json();
                const items = cartData.items || [];

                // Fetch product details for each cart item
                if (items.length > 0) {
                    const productsResponse = await fetch('/api/products');
                    if (productsResponse.ok) {
                        const products = await productsResponse.json();

                        // Merge cart items with product details
                        const enrichedItems = items.map(cartItem => {
                            const product = products.find(p => String(p.id) === String(cartItem.product_id));
                            return product ? { ...product, quantity: cartItem.quantity } : null;
                        }).filter(Boolean);

                        setCartItems(enrichedItems);
                    } else {
                        setCartItems(items); // Fallback to cart data only
                    }
                } else {
                    setCartItems([]);
                }
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
            const response = await fetch('/api/cart/guest', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ product_id: String(product.id), quantity: 1 }),
            });
            if (response.ok) {
                await fetchCart();
                setIsOpen(true);
            }
        } catch (error) {
            console.error('Failed to add to cart:', error);
        }
    };

    const removeFromCart = async (productId) => {
        try {
            // Delete specific item by clearing cart and re-adding other items
            // Note: This is a workaround since cart-service doesn't have a delete item endpoint
            const itemsToKeep = cartItems.filter(item => String(item.id) !== String(productId));

            // Clear the cart
            await fetch('/api/cart/guest', { method: 'DELETE' });

            // Re-add remaining items
            for (const item of itemsToKeep) {
                await fetch('/api/cart/guest', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ product_id: String(item.id), quantity: item.quantity }),
                });
            }

            await fetchCart();
        } catch (error) {
            console.error('Failed to remove from cart:', error);
        }
    };

    const checkout = async () => {
        try {
            // Build cartItems array from current cart (camelCase for Java DTO)
            const checkoutItems = cartItems.map(item => ({
                productId: String(item.id),
                quantity: item.quantity
            }));

            // Call checkout-service with userId and cartItems (camelCase)
            const response = await fetch('/api/checkout', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    userId: 'guest',
                    cartItems: checkoutItems
                }),
            });
            if (response.ok) {
                const data = await response.json();
                // Store the checkout response with transaction details
                setCheckoutResponse(data);
                // Clear cart after successful checkout
                await fetch('/api/cart/guest', { method: 'DELETE' });
                setCartItems([]);
                setIsOpen(false);
                return true;
            } else {
                const errorData = await response.json();
                console.error('Checkout failed:', errorData);
            }
        } catch (e) {
            console.error("Checkout error", e);
        }
        return false;
    };

    return (
        <CartContext.Provider value={{ cartItems, addToCart, removeFromCart, checkout, isOpen, setIsOpen, fetchCart, isLoading, checkoutResponse, setCheckoutResponse }}>
            {children}
        </CartContext.Provider>
    );
};
