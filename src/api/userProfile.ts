import { User } from "../@types/user";
import { Team } from "../@types/team";
import { ErrorResponse } from "../@types/error";

export interface TeamUserProfileData {
    [key: string]: TeamUserProfileResult, // myTeams, allTeams
}

export interface TeamUserProfileResult {
    teams: Team[]
    count: number
}


export const getUserProfile = async (principalName: string, token?: string): Promise<User | ErrorResponse> => {
    const accessToken = localStorage.getItem('access_token');

    // TODO: should not need this logic. Should be able to use principalName as is
    if (principalName.endsWith('@ssb.no')) {
        principalName = principalName.replace('@ssb.no', '');
    }

    return fetch(`/api/userProfile/${principalName}@ssb.no`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${accessToken || token}`
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

export const getUserTeamsWithGroups = async (principalName: string): Promise<Team[] | ErrorResponse> => {
    const accessToken = localStorage.getItem('access_token');

    // TODO: should not need this logic. Should be able to use principalName as is
    if (principalName.endsWith('@ssb.no')) {
        principalName = principalName.replace('@ssb.no', '');
    }

    return fetch(`/api/userProfile/${principalName}@ssb.no/team`, {
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
    }).then(data => data as Team[])
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