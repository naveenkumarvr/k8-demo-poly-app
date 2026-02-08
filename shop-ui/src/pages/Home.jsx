import React, { useEffect, useState } from 'react';
import ProductGrid from '../components/ProductGrid';

export default function Home() {
    const [products, setProducts] = useState([]);
    const [ads, setAds] = useState([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const [productsRes, adsRes] = await Promise.allSettled([
                    fetch('/api/products'),
                    fetch('/api/ads')
                ]);

                if (productsRes.status === 'fulfilled' && productsRes.value.ok) {
                    const data = await productsRes.value.json();
                    setProducts(data.products || data);
                }

                if (adsRes.status === 'fulfilled' && adsRes.value.ok) {
                    const data = await adsRes.value.json();
                    setAds(data.ads || []);
                }

            } catch (error) {
                console.error('Error fetching data:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, []);

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[50vh]">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    return (
        <div className="space-y-8">
            {/* Hero / Ad Wrapper */}
            {ads.length > 0 && (
                <div className="bg-gradient-to-r from-blue-600 to-indigo-700 rounded-2xl p-8 text-white shadow-lg mb-8">
                    <h2 className="text-3xl font-bold mb-2">Special Offers</h2>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {ads.map((ad, i) => (
                            <a key={i} href={ad.redirect_url} className="block bg-white/10 hover:bg-white/20 p-4 rounded-lg transition-colors">
                                <span className="font-semibold block text-lg">{ad.text}</span>
                            </a>
                        ))}
                    </div>
                </div>
            )}

            <div>
                <h1 className="text-2xl font-bold text-gray-900 mb-6">Featured Products</h1>
                <ProductGrid products={products} />
            </div>
        </div>
    );
}
