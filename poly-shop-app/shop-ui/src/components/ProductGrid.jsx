import React from 'react';
import { useCart } from '../context/CartContext.jsx';
import { Plus } from 'lucide-react';

export default function ProductGrid({ products }) {
    const { addToCart } = useCart();

    if (!products || products.length === 0) {
        return <div className="text-center py-10 text-gray-500">No products available</div>;
    }

    return (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {products.map((product) => (
                <div key={product.id} className="group bg-white rounded-xl shadow-sm border border-gray-100 hover:shadow-md transition-shadow overflow-hidden flex flex-col">
                    <div className="aspect-square bg-gray-50 relative overflow-hidden">
                        <img
                            src={product.image_url}
                            alt={product.name}
                            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                        />
                    </div>
                    <div className="p-4 flex flex-col flex-1">
                        <h3 className="font-medium text-gray-900 mb-1">{product.name}</h3>
                        <p className="text-gray-500 text-sm line-clamp-2 flex-1 mb-3">{product.description}</p>
                        <div className="flex items-center justify-between mt-auto">
                            <span className="font-bold text-lg text-blue-600">
                                ${typeof product.price === 'number' ? product.price.toFixed(2) : '0.00'}
                            </span>
                            <button
                                onClick={() => addToCart(product)}
                                className="p-2 bg-gray-900 text-white rounded-full hover:bg-gray-800 transition-colors"
                                aria-label="Add to cart"
                            >
                                <Plus className="w-5 h-5" />
                            </button>
                        </div>
                    </div>
                </div>
            ))}
        </div>
    );
}
