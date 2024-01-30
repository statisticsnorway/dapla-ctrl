import React, { FC, ReactNode, createContext, useState, useCallback } from 'react';
import { jwtRegex } from '../utils/regex';

interface DaplaCtrlContextType {
    breadcrumbUserProfileDisplayName: string | null;
    setBreadcrumbUserProfileDisplayName: (breadcrumbUserProfileDisplayName: string | null) => void;
    isLoggedIn: boolean;
    login: (token: string, userProfile: object) => void;
    logout: () => void;
}

const DaplaCtrlContext = createContext<DaplaCtrlContextType>({
    breadcrumbUserProfileDisplayName: "",
    setBreadcrumbUserProfileDisplayName: () => { },
    isLoggedIn: false,
    login: () => { },
    logout: () => { }
});

const DaplaCtrlProvider: FC<{ children: ReactNode }> = ({ children }) => {
    const [breadcrumbUserProfileDisplayName, setBreadcrumbUserProfileDisplayName] = useState<string | null>(null);
    const [isLoggedIn, setIsLoggedIn] = useState<boolean>(false);

    const checkLoginState = useCallback(() => {
        const accessToken = localStorage.getItem('access_token');
        const userProfile = localStorage.getItem('userProfile');
        const isTokenValid = accessToken !== null && jwtRegex.test(accessToken);
        const isUserProfileSet = userProfile !== null;

        setIsLoggedIn(isTokenValid && isUserProfileSet);
    }, []);

    const login = useCallback((token: string, userProfile: object) => {
        localStorage.setItem('access_token', token);
        localStorage.setItem('userProfile', JSON.stringify(userProfile));
        checkLoginState();
    }, [checkLoginState]);

    const logout = useCallback(() => {
        localStorage.removeItem('access_token');
        localStorage.removeItem('userProfile');
        checkLoginState();
    }, [checkLoginState]);

    // Check the login state initially and whenever the relevant localStorage items change
    React.useEffect(() => {
        checkLoginState();
        window.addEventListener('storage', checkLoginState);
        return () => {
            window.removeEventListener('storage', checkLoginState);
        };
    }, [checkLoginState]);

    return (
        <DaplaCtrlContext.Provider value={{ breadcrumbUserProfileDisplayName, setBreadcrumbUserProfileDisplayName, isLoggedIn, login, logout }}>
            {children}
        </DaplaCtrlContext.Provider>
    );
};

export { DaplaCtrlContext, DaplaCtrlProvider };
