import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { verifyKeycloakToken } from '../api/VerifyKeycloakToken';

type ProtectedRouteProps = {
    children: React.ReactNode;
};

export const ProtectedRoute = ({ children }: ProtectedRouteProps) => {
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const navigate = useNavigate();
    const from = location.pathname;

    useEffect(() => {
        const token = localStorage.getItem('token');
        if (token) {
            verifyKeycloakToken(token).then(isValid => {
                setIsAuthenticated(isValid);
                if (!isValid) {
                    localStorage.removeItem('token');
                    navigate('/login', { state: { from: from } });
                }
            });
        } else {
            navigate('/login', { state: { from: from } });
        }
    }, [navigate]);

    return isAuthenticated ? children : null;
};
