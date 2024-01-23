import styles from './login.module.scss';

import secureLocalStorage from 'react-secure-storage';

import { Title, Input, Link } from "@statisticsnorway/ssb-component-library";
import { useEffect, useState } from "react";
import { verifyKeycloakToken } from "../../api/VerifyKeycloakToken";
import { useLocation, useNavigate } from 'react-router-dom';

const jwtRegex = /^[A-Za-z0-9-_=]+\.[A-Za-z0-9-_=]+\.?[A-Za-z0-9-_.+/=]*$/;

export default function Login() {
    const [error, setError] = useState(false);
    const [value, setValue] = useState("");

    const navigate = useNavigate();
    const location = useLocation();
    const from = location.state?.from || '/';

    useEffect(() => {
        const storedAccessToken = secureLocalStorage.getItem('access_token') as string;

        if (storedAccessToken && jwtRegex.test(storedAccessToken)) {
            verifyKeycloakToken(storedAccessToken).then(isValid => {
                if (isValid) {
                    navigate(from);
                }
            });
        }
    }, [navigate]);

    useEffect(() => {
        const validateAccessToken = (accessToken: string) => {
            // Check if the token matches the JWT pattern
            if (!jwtRegex.test(accessToken)) {
                return Promise.resolve(false);
            }

            // Check if the token is invalid
            return verifyKeycloakToken(accessToken).then(isValid => {
                if (!isValid) {
                    return false;
                }
                setValue(accessToken);
                return true;
            });
        };

        if (!value) {
            setError(false);
        } else {
            validateAccessToken(value).then(isValidToken => {
                if (isValidToken) {
                    secureLocalStorage.setItem('access_token', value)
                    navigate(from);
                }
                setError(!isValidToken);
            });
        }
    }, [value, from]);

    const handleInputChange = (input: string) => {
        setValue(input);
    };

    return (
        <div className={styles.loginContainer}>
            <Title size={1}>Logg inn med token</Title>
            <span>
                Trykk <Link isExternal={true} href="https://httpbin-fe.staging-bip-app.ssb.no/bearer">her</Link> for å hente keycloak token
            </span>
            <Input
                label="Lim inn keycloak token"
                placeholder="Keycloak token" 
                type="password"
                value={value} 
                handleChange={handleInputChange}
                error={error} 
                errorMessage="Invalid keycloak token" />
        </div>
    )
}
