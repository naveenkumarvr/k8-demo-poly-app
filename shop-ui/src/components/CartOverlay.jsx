import React from 'react';
import { useCart } from '../context/CartContext.jsx';
import { X, ShoppingBag, Trash2 } from 'lucide-react';

export default function CartOverlay() {
    const { cartItems, isOpen, setIsOpen, checkout, removeFromCart } = useCart();

    if (!isOpen) return null;

    const total = cartItems.reduce((sum, item) => sum + ((typeof item.price === 'number' ? item.price : 0) * item.quantity), 0);

    return (
        <div className="fixed inset-0 z-50 flex justify-end">
            {/* Backdrop */}
            <div
                className="absolute inset-0 bg-black/50 transition-opacity"
                onClick={() => setIsOpen(false)}
            />

            {/* Drawer */}
            <div className="relative w-full max-w-md bg-white shadow-xl h-full flex flex-col transform transition-transform duration-300">
                <div className="p-4 border-b flex items-center justify-between">
                    <h2 className="text-lg font-semibold flex items-center gap-2">
                        <ShoppingBag className="w-5 h-5" />
                        Your Cart
                    </h2>
                    <button
                        onClick={() => setIsOpen(false)}
                        className="p-2 hover:bg-gray-100 rounded-full"
                    >
                        <X className="w-5 h-5" />
                    </button>
                </div>

                <div className="flex-1 overflow-y-auto p-4 space-y-4">
                    {cartItems.length === 0 ? (
                        <div className="text-center text-gray-500 mt-10">
                            Your cart is empty
                        </div>
                    ) : (
                        cartItems.map((item, idx) => (
                            <div key={idx} className="flex gap-4 border-b pb-4">
                                <div className="w-20 h-20 bg-gray-100 rounded-md overflow-hidden flex-shrink-0">
                                    <img src={item.image_url || item.picture} alt={item.name} className="w-full h-full object-cover" />
                                </div>
                                <div className="flex-1 min-w-0">
                                    <h3 className="font-medium truncate">{item.name}</h3>
                                    <p className="text-gray-500 text-sm">Qty: {item.quantity}</p>
                                    <p className="font-semibold mt-1">
                                        ${typeof item.price === 'number' ? item.price.toFixed(2) : '0.00'}
                                    </p>
                                </div>
                                <button
                                    onClick={() => removeFromCart(item.id)}
                                    className="self-start p-2 text-red-500 hover:bg-red-50 rounded-full transition-colors"
                                    aria-label="Remove item"
                                >
                                    <Trash2 className="w-5 h-5" />
                                </button>
                            </div>
                        ))
                    )}
                </div>

                <div className="p-4 border-t bg-gray-50">
                    <div className="flex justify-between mb-4 text-lg font-bold">
                        <span>Total</span>
                        <span>${total.toFixed(2)}</span>
                    </div>
                    <button
                        onClick={checkout}
                        disabled={cartItems.length === 0}
                        className="w-full py-3 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                    >
                        Checkout
                    </button>
                </div>
            </div>
        </div>
    );
}
