import { Input } from "@statisticsnorway/ssb-component-library";
import { useEffect, useState } from "react";
import { verifyKeycloakToken } from "../api/VerifyKeycloakToken";
import { useLocation, useNavigate } from 'react-router-dom';

const jwtRegex = /^[A-Za-z0-9-_=]+\.[A-Za-z0-9-_=]+\.?[A-Za-z0-9-_.+/=]*$/;

export function Login() {
    const [error, setError] = useState(false);
    const [value, setValue] = useState("");

    const navigate = useNavigate();
    const location = useLocation();
    const from = location.state?.from || '/';

    useEffect(() => {
        // Check for token in local storage
        const storedToken = localStorage.getItem("token");

        // If a token is found, verify it
        if (storedToken && jwtRegex.test(storedToken)) {
            verifyKeycloakToken(storedToken).then(isValid => {
                if (isValid) {
                    // If the token is valid, redirect to home page
                    navigate(from);
                }
            });
        }
    }, [navigate]);

    useEffect(() => {
        if (!value) {
            setError(false);
        }

        const checkToken = async () => {
            const isValidToken = await validateToken(value);
            if (isValidToken) {
                localStorage.setItem("token", value);
                navigate(from);
            }
            setError(!isValidToken);
        }

        if (value) {
            checkToken();
        }
    }, [value, from]);

    const validateToken = async (token: string) => {
        // Check if the token matches the JWT pattern
        if (!jwtRegex.test(token)) {
            return false;
        }

        // Check if the token is valid
        if (!await verifyKeycloakToken(token)) {
            return false;
        }

        setValue(token);
        return true;
    };

    const handleInputChange = (input: string) => {
        setValue(input);
    };

    return (
        <>
            <div className="container">
                <Input label="Password" placeholder={"Keycloak token"} value={value} handleChange={handleInputChange} error={error} errorMessage="Invalid keycloak token" />
            </div>
        </>
    )
}
