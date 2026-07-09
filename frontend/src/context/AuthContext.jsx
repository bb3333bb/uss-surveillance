import React, { createContext, useState, useEffect, useContext } from 'react';
import axios from 'axios';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [token, setToken] = useState(sessionStorage.getItem('jwt') || null);
  const [user, setUser] = useState(sessionStorage.getItem('user') || null);
  const [role, setRole] = useState(sessionStorage.getItem('role') || null);
  const [loading, setLoading] = useState(true);

  // Configure global axios interceptor to attach Bearer Token if present
  useEffect(() => {
    const requestInterceptor = axios.interceptors.request.use(
      (config) => {
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    return () => {
      axios.interceptors.request.eject(requestInterceptor);
    };
  }, [token]);

  useEffect(() => {
    // Process OIDC URL callback parameters if returning from IDP login
    const params = new URLSearchParams(window.location.search);
    const callbackToken = params.get('token');
    const callbackUser = params.get('user');
    const callbackRole = params.get('role');

    if (callbackToken && callbackUser && callbackRole) {
      sessionStorage.setItem('jwt', callbackToken);
      sessionStorage.setItem('user', callbackUser);
      sessionStorage.setItem('role', callbackRole);

      setToken(callbackToken);
      setUser(callbackUser);
      setRole(callbackRole);

      // Clean URL params after successful parse
      window.history.replaceState({}, document.title, window.location.pathname);
    }
    setLoading(false);
  }, []);

  const login = () => {
    // Redirect browser to Go gateway OIDC initiator endpoint
    window.location.href = 'http://localhost:8080/api/auth/login';
  };

  const logout = () => {
    sessionStorage.removeItem('jwt');
    sessionStorage.removeItem('user');
    sessionStorage.removeItem('role');

    setToken(null);
    setUser(null);
    setRole(null);
  };

  const value = {
    token,
    user,
    role,
    loading,
    login,
    logout,
    isAuthenticated: !!token,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be executed within an AuthProvider');
  }
  return context;
};
