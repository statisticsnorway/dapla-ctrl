import { useContext, useEffect, useState } from 'react';
import { useNavigate, Outlet } from 'react-router-dom';
import { validateKeycloakToken } from '../api/validateKeycloakToken';
import { DaplaCtrlContext } from '../provider/DaplaCtrlProvider';

const ProtectedRoute = () => {
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const navigate = useNavigate();
    const from = location.pathname;
    const { logout } = useContext(DaplaCtrlContext);

    useEffect(() => {
        if (localStorage.getItem('userProfile') === null) {
            logout();
            navigate('/login', { state: { from: from } });
            return;
        }

        validateKeycloakToken().then(isValid => {
            setIsAuthenticated(isValid);
            if (!isValid) {
                logout();
                navigate('/login', { state: { from: from } });
            }
        });
    }, [navigate]);

    return isAuthenticated ? <Outlet /> : null;
};

export default ProtectedRoute;