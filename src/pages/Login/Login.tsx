import styles from './login.module.scss'

import Cookies from 'js-cookie'

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
        const storedToken = Cookies.get('token');

        if (storedToken && jwtRegex.test(storedToken)) {
            verifyKeycloakToken(storedToken).then(isValid => {
                if (isValid) {
                    navigate(from);
                }
            });
        }
    }, [navigate]);

    useEffect(() => {
        const validateToken = (token: string) => {
            // Check if the token matches the JWT pattern
            if (!jwtRegex.test(token)) {
                return Promise.resolve(false);
            }

            // Check if the token is invalid
            return verifyKeycloakToken(token).then(isValid => {
                if (!isValid) {
                    return false;
                }
                setValue(token);
                return true;
            });
        };

        if (!value) {
            setError(false);
        } else {
            validateToken(value).then(isValidToken => {
                if (isValidToken) {
                    Cookies.set('token', value, { expires: 7, secure: true, sameSite: 'strict'})
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
                value={value} 
                handleChange={handleInputChange} 
                error={error} 
                errorMessage="Invalid keycloak token" />
        </div>
    )
}