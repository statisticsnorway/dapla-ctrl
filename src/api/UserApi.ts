
export interface User {
    principalName: string
    azureAdId: string
    displayName: string
    firstName: string
    lastName: string
    email: string
    manager?: User
    photo?: string
}

export const getUserProfile = async (accessToken: string): Promise<User> => {
    return fetch('/api/userProfile', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${accessToken}`
        }
    }).then(response => {
        if (!response.ok) {
            console.error('Request failed with status:', response.status);
            throw new Error('Request failed');
        }
        return response.json();
    }).then(data => data as User)
        .catch(error => {
            console.error('Error during fetching userProfile:', error);
            throw error;
        });
};

export const getUserProfileFallback = (accessToken: string): User => {
    const jwt = JSON.parse(atob(accessToken.split('.')[1]));
    return {
        principalName: jwt.upn,
        azureAdId: jwt.oid, // not the real azureAdId, this is actually keycloaks oid
        displayName: jwt.name,
        firstName: jwt.given_name,
        lastName: jwt.family_name,
        email: jwt.email
    };
};