import Cookies from 'js-cookie';

import { useEffect, useState } from 'react';
import { useNavigate, Outlet } from 'react-router-dom';
import { verifyKeycloakToken } from '../api/VerifyKeycloakToken';

export const ProtectedRoute = () => {
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const navigate = useNavigate();
    const from = location.pathname;

    useEffect(() => {
        const token = Cookies.get('token');
        
        if (token) {
            verifyKeycloakToken(token).then(isValid => {
                setIsAuthenticated(isValid);
                if (!isValid) {
                    Cookies.remove('token');
                    navigate('/login', { state: { from: from } });
                }
            });
        } else {
            navigate('/login', { state: { from: from } });
        }
    }, [navigate]);

    return isAuthenticated ? <Outlet /> : null;
};
