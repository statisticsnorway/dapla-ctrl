import { useEffect, useState } from 'react';
import { useNavigate, Outlet } from 'react-router-dom';
import { verifyKeycloakToken } from '../api/VerifyKeycloakToken';

export const ProtectedRoute = () => {
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const navigate = useNavigate();
    const from = location.pathname;

    useEffect(() => {
        if (localStorage.getItem('userProfile') === null) {
            localStorage.removeItem('access_token');
            navigate('/login', { state: { from: from } });
            return;
        }

        verifyKeycloakToken().then(isValid => {
            setIsAuthenticated(isValid);
            if (!isValid) {
                localStorage.removeItem('access_token');
                localStorage.removeItem('userProfile');
                navigate('/login', { state: { from: from } });
            }
        });
    }, [navigate]);

    return isAuthenticated ? <Outlet /> : null;
};
