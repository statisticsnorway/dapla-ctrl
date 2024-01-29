export interface TeamOverviewData {
    [key: string]: Root,
}

export interface Root {
    _embedded: Embedded
    count: number
}

export interface Embedded {
    teams: Team[]
}

export interface Team {
    uniformName: string
    displayName: string
    teamUserCount: number
    manager: Manager
}

export interface Manager {
    principalName: string
    displayName: string
}

export interface TeamOverviewError {
    success: boolean
    error: Error
}

export interface Error {
    code: string
    message: string
}

export const getTeamOverview = async (): Promise<TeamOverviewData | TeamOverviewError> => {
    const accessToken = localStorage.getItem('access_token');

    try {
        const response = await fetch(`/api/teamOverview`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${accessToken}`
            }
        });
        if (!response.ok) {
            const errorData = await response.json();
            return errorData as TeamOverviewError;
        }
        const data = await response.json();
        return data as TeamOverviewData;
    } catch (error) {
        console.error('Error during fetching teams:', error);
        throw new Error('Error fetching teams');
    }
};