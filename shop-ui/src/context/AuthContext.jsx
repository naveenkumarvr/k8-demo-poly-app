import React, { createContext, useContext, useState, useEffect } from 'react';

const AuthContext = createContext();

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};

const GUEST_USER = 'guest';
const DEMO_USER = 'user-1';
const DEMO_PASSWORD = 'password';

export const AuthProvider = ({ children }) => {
    const [userId, setUserId] = useState(() => {
        return localStorage.getItem('polyshop-user-id') || null;
    });
    const [isGuest, setIsGuest] = useState(() => {
        return localStorage.getItem('polyshop-is-guest') === 'true';
    });
    const [error, setError] = useState(null);

    const loginAsGuest = () => {
        setUserId(GUEST_USER);
        setIsGuest(true);
        setError(null);
        localStorage.setItem('polyshop-user-id', GUEST_USER);
        localStorage.setItem('polyshop-is-guest', 'true');
    };

    const loginAsUser = (username, password) => {
        if (username !== DEMO_USER) {
            setError('Invalid username. Try user-1.');
            return false;
        }
        if (password !== DEMO_PASSWORD) {
            setError('Invalid password. Try password.');
            return false;
        }
        setUserId(DEMO_USER);
        setIsGuest(false);
        setError(null);
        localStorage.setItem('polyshop-user-id', DEMO_USER);
        localStorage.setItem('polyshop-is-guest', 'false');
        return true;
    };

    const logout = () => {
        setUserId(null);
        setIsGuest(false);
        setError(null);
        localStorage.removeItem('polyshop-user-id');
        localStorage.removeItem('polyshop-is-guest');
    };

    const isAuthenticated = userId !== null;

    const value = {
        userId,
        isGuest,
        isAuthenticated,
        loginAsGuest,
        loginAsUser,
        logout,
        error,
        setError,
    };

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
