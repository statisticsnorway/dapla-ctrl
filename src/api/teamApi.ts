export interface ApiResponse {
    success: boolean;
    data: Data;
}

interface Data {
    _embedded: Embedded;
    _links: Links;
    count: number;
}

interface Embedded {
    teams: Team[];
}

interface Team {
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

interface Links {
    self: Link;
}


export const getAllTeams = (token: string): Promise<ApiResponse> => {
    return fetch('/api/teams', {
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
    }).then(data => data as ApiResponse)
        .catch(error => {
            console.error('Error during fetching teams:', error);
            throw error;
        });
};

