import { Team } from '../@types/team';
import { ErrorResponse } from '../@types/error';

export interface TeamOverviewData {
    [key: string]: TeamOverviewResult, // myTeams, allTeams
}

export interface TeamOverviewResult {
    teams: Team[]
    count: number
}

export const getTeamOverview = async (): Promise<TeamOverviewData | ErrorResponse> => {
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
            return errorData as ErrorResponse;
        }
        const data = await response.json();
        return data as TeamOverviewData;
    } catch (error) {
        console.error('Error during fetching teams:', error);
        throw new Error('Error fetching teams');
    }
};