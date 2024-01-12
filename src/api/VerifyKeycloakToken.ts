export const verifyKeycloakToken = async (token: string): Promise<boolean> => {
    try {
        // Send a request to the server to verify the token
        const response = await fetch('/verify-token', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ token: token })
        });

        // Check if the response status is 200 and return the result
        return response.status === 200;
    } catch (error) {
        console.error('Error during token validation:', error);
        return false;
    }
};
