import styles from './login.module.scss'

import { Title, Input, Link } from "@statisticsnorway/ssb-component-library";
import { useEffect, useState } from "react";
import { verifyKeycloakToken } from "../../api/VerifyKeycloakToken";
import { useLocation, useNavigate } from 'react-router-dom';
import { jwtRegex } from "../../utils/regex";
import { getUserProfile } from "../../api/UserProfile";

export default function Login() {
    const [error, setError] = useState(false);
    const [value, setValue] = useState("");

    const navigate = useNavigate();
    const location = useLocation();
    const from = location.state?.from || '/';

    useEffect(() => {
        const storedToken = localStorage.getItem("token");

        if (storedToken && jwtRegex.test(storedToken)) {
            verifyKeycloakToken(storedToken).then(isValid => {
                if (isValid) {
                    navigate(from);
                }
            });
        }
    }, [navigate]);

    useEffect(() => {
        const validateToken = async (token: string) => {
            // Check if the token matches the JWT pattern
            if (!jwtRegex.test(token)) return false;

            // Check if the token is invalid
            const isValid = await verifyKeycloakToken(token);
            if (!isValid) return false;

            setValue(token);
            const userProfile = await getUserProfile(token);
            localStorage.setItem("userProfile", JSON.stringify(userProfile));
            return true;
        };

        if (!value) {
            setError(false);
        } else {
            validateToken(value).then(isValidToken => {
                if (isValidToken) {
                    localStorage.setItem("token", value);
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
