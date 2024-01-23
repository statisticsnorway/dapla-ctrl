
export interface UserData {
    principalName: string
    azureAdId: string
    displayName: string
    firstName: string
    lastName: string
    email: string,
    manager: any
    photo: string
}


export const getUserProfile = async (token: string): Promise<UserData> => {
    return fetch('/api/userProfile', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        }
    }).then(response => {
        if (!response.ok) {
            console.error('Request failed with status:', response.status);
            throw new Error('Request failed');
        }
        return response.json();
    }).then(data => data as UserData)
        .catch(error => {
            console.error('Error during fetching teams:', error);
            throw error;
        });
};