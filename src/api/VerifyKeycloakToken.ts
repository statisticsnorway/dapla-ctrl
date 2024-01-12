export const verifyKeycloakToken = (token: string): Promise<boolean> => {
    return fetch('/verify-token', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ token })
    }).then(response => {
        if (!response.ok) {
            console.error('Token verification failed with status:', response.status);
            return false;
        }
        return response.status === 200;
    }).catch(error => {
        console.error('Error during token validation:', error);
        return false;
    });
};
