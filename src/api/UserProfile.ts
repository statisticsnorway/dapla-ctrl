
export interface UserData {
    principalName: string
    azureAdId: string
    displayName: string
    firstName: string
    lastName: string
    email: string,
    manager: any
    photo: string | null
}


export const getUserProfile = async (accessToken: string): Promise<UserData> => {

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
    }).then(data => data as UserData)
        .catch(error => {
            console.error('Error during fetching userProfile:', error);
            throw error;
        });
};

export const getUserProfileFallback = (accessToken: string): UserData => {
    var jwt = JSON.parse(atob(accessToken.split('.')[1]));
    return {
        principalName: jwt.upn,
        azureAdId: jwt.oid,
        displayName: jwt.name,
        firstName: jwt.given_name,
        lastName: jwt.family_name,
        email: jwt.email,
        manager: null,
        photo: null
    };
};