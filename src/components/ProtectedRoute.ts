import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { verifyKeycloakToken } from '../api/VerifyKeycloakToken';

export const ProtectedRoute = ({ children }) => {
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const navigate = useNavigate();
    const from = location.pathname;

    useEffect(() => {
        const checkAuthentication = async () => {
            const token = localStorage.getItem('token');
            if (token) {
                const isValid = await verifyKeycloakToken(token);
                setIsAuthenticated(isValid);
                if (!isValid) {
                    localStorage.removeItem('token');
                    navigate('/login', { state: { from: from } });
                }
            } else {
                navigate('/login', { state: { from: from } });
            }
        };

        checkAuthentication();
    }, [navigate]);

    return isAuthenticated ? children : null;
};
