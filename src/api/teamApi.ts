import { getTokenCookie } from "../auth";

export interface TeamApiResponse {
    success: boolean;
    data: Team[];
}

export interface Team {
    uniformName: string;
    displayName: string;
    _links: TeamLinks;
}

interface TeamLinks {
    self: Link;
}

interface Link {
    href: string;
    templated?: boolean;
}


export const getAllTeams = (): Promise<TeamApiResponse> => {
    return fetch('/api/teams', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${getTokenCookie()}`
        }
    }).then(response => {
        if (!response.ok) {
            console.error('Request failed with status:', response.status);
            throw new Error('Request failed');
        }
        return response.json();
    }).then(data => data as TeamApiResponse)
        .catch(error => {
            console.error('Error during fetching teams:', error);
            throw error;
        });
};

