import styles from './login.module.scss'

import { useEffect, useState } from "react";
import { useLocation, useNavigate } from 'react-router-dom';

import { Title, Input, Link } from "@statisticsnorway/ssb-component-library";

import { verifyKeycloakToken } from "../../api/VerifyKeycloakToken";
import { getUserProfile, getUserProfileFallback } from "../../api/UserApi";
import { jwtRegex } from "../../utils/regex";

export default function Login() {
    const [error, setError] = useState(false);
    const [value, setValue] = useState("");

    const navigate = useNavigate();
    const location = useLocation();
    const from = location.state?.from || '/';

    useEffect(() => {
        const storedAccessToken = localStorage.getItem('access_token') as string;

        if (storedAccessToken && jwtRegex.test(storedAccessToken)) {
            verifyKeycloakToken(storedAccessToken).then(isValid => {
                if (isValid) {
                    navigate(from);
                }
            });
        }
    }, [navigate]);

    useEffect(() => {
        const validateToken = async (accessToken: string) => {
            // Check if the token matches the JWT pattern
            if (!jwtRegex.test(accessToken)) return false;

            // Check if the token is invalid
            const isValid = await verifyKeycloakToken(accessToken);
            if (!isValid) return false;
            setValue(accessToken);

            try {
                const userProfile = await getUserProfile(accessToken);
                localStorage.setItem("userProfile", JSON.stringify(userProfile));
            } catch (error) {
                console.error("Could not fetch user profile, using fallback", error);
                const userProfile = getUserProfileFallback(accessToken);
                localStorage.setItem("userProfile", JSON.stringify(userProfile));
            }

            return true;
        };

        if (!value) {
            setError(false);
        } else {
            validateToken(value).then(isValidAccessToken => {
                if (isValidAccessToken) {
                    localStorage.setItem("access_token", value);
                    navigate(from);
                }
                setError(!isValidAccessToken);
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
