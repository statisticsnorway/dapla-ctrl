import Cookies from 'js-cookie';

import { useEffect, useState } from 'react';
import { useNavigate, Outlet } from 'react-router-dom';
import { verifyKeycloakToken } from '../api/VerifyKeycloakToken';

export const ProtectedRoute = () => {
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const navigate = useNavigate();
    const from = location.pathname;

    useEffect(() => {
        verifyKeycloakToken().then(isValid => {
            setIsAuthenticated(isValid);
            if (!isValid) {
                Cookies.remove('access_token', { secure: true, sameSite: 'strict' });
                navigate('/login', { state: { from: from } });
            }
        });
    }, [navigate]);

    return isAuthenticated ? <Outlet /> : null;
};
