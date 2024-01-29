
export interface User {
    principal_name: string
    azure_ad_id: string
    display_name: string
    first_name: string
    last_name: string
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
        principal_name: jwt.upn,
        azure_ad_id: jwt.oid, // not the real azureAdId, this is actually keycloaks oid
        display_name: jwt.name,
        first_name: jwt.given_name,
        last_name: jwt.family_name,
        email: jwt.email
    };
};