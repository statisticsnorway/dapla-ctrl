import { getRequest } from "./Requests"

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

export const getAllTeams = (): Promise<Team[]> => {
    const accessToken = localStorage.getItem('access_token');

	return getRequest('/api/teams', accessToken).then(response => {
        if (!response.ok) {
            console.error('Request failed with status:', response.status);
            throw new Error('Request failed');
        }
        return response.json();
    }).then(data => data as Team[])
        .catch(error => {
            console.error('Error during fetching teams:', error);
            throw error;
        });
};

