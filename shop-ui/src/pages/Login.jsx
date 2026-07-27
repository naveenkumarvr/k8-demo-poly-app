import React, { useState } from 'react';
import { Package, User, UserCircle } from 'lucide-react';

export default function Login({ onLogin }) {
    const [username, setUsername] = useState('user-1');
    const [password, setPassword] = useState('');
    const [error, setError] = useState(null);

    const handleGuest = () => {
        onLogin({ type: 'guest' });
    };

    const handleUserLogin = (e) => {
        e.preventDefault();
        setError(null);
        const success = onLogin({ type: 'user', username, password });
        if (!success) {
            setError('Invalid username or password. Use user-1 / password.');
        }
    };

    return (
        <div className="min-h-screen bg-gray-50 flex flex-col">
            <header className="bg-white shadow-sm">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center gap-2 text-blue-600">
                    <Package className="w-8 h-8" />
                    <span className="font-bold text-xl tracking-tight">PolyShop</span>
                </div>
            </header>

            <main className="flex-grow flex items-center justify-center px-4 py-12">
                <div className="w-full max-w-md bg-white rounded-2xl shadow-lg p-8 space-y-6">
                    <div className="text-center space-y-2">
                        <h1 className="text-2xl font-bold text-gray-900">Welcome to PolyShop</h1>
                        <p className="text-gray-500">Sign in to start shopping</p>
                    </div>

                    <button
                        onClick={handleGuest}
                        className="w-full py-3 px-4 bg-gray-100 hover:bg-gray-200 text-gray-900 rounded-lg font-medium flex items-center justify-center gap-2 transition-colors"
                    >
                        <UserCircle className="w-5 h-5" />
                        Continue as Guest
                    </button>

                    <div className="relative">
                        <div className="absolute inset-0 flex items-center">
                            <div className="w-full border-t border-gray-200"></div>
                        </div>
                        <div className="relative flex justify-center text-sm">
                            <span className="px-2 bg-white text-gray-500">or</span>
                        </div>
                    </div>

                    <form onSubmit={handleUserLogin} className="space-y-4">
                        <div>
                            <label htmlFor="username" className="block text-sm font-medium text-gray-700 mb-1">
                                Username
                            </label>
                            <div className="relative">
                                <User className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                                <input
                                    id="username"
                                    type="text"
                                    value={username}
                                    onChange={(e) => setUsername(e.target.value)}
                                    className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
                                    placeholder="user-1"
                                />
                            </div>
                        </div>

                        <div>
                            <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
                                Password
                            </label>
                            <input
                                id="password"
                                type="password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
                                placeholder="password"
                            />
                        </div>

                        {error && (
                            <div className="text-sm text-red-600 bg-red-50 px-4 py-2 rounded-lg">
                                {error}
                            </div>
                        )}

                        <button
                            type="submit"
                            className="w-full py-3 px-4 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors"
                        >
                            Sign In
                        </button>

                        <p className="text-xs text-center text-gray-500">
                            Demo credentials: <span className="font-medium">user-1</span> / <span className="font-medium">password</span>
                        </p>
                    </form>
                </div>
            </main>

            <footer className="bg-white border-t py-6">
                <div className="max-w-7xl mx-auto px-4 text-center text-gray-500 text-sm">
                    &copy; {new Date().getFullYear()} PolyShop Microservices Demo.
                </div>
            </footer>
        </div>
    );
}
