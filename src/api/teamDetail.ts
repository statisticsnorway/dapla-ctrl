import { Team } from '../@types/team'
import { User } from '../@types/user'
import { ErrorResponse } from '../@types/error'

export interface TeamDetailData {
  [key: string]: TeamDetailResult // myTeams, allTeams
}

export interface TeamDetailResult {
  teamInfo: Team
  teamUsers: User[]
  count: number
}

export const getTeamDetail = async (teamId: string): Promise<TeamDetailData | ErrorResponse> => {
  const accessToken = localStorage.getItem('access_token')

  try {
    const response = await fetch(`/api/teamDetail/${teamId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${accessToken}`,
      },
    })
    if (!response.ok) {
      const errorData = await response.json()
      return errorData as ErrorResponse
    }
    const data = await response.json()
    return data as TeamDetailData
  } catch (error) {
    console.error('Error during fetching teams:', error)
    throw new Error('Error fetching teams')
  }
}
