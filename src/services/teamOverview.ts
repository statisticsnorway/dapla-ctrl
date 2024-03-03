import { ApiError, fetchAPIData } from '../utils/services'
import { flattenEmbedded } from '../utils/utils'
const DAPLA_TEAM_API_URL = import.meta.env.VITE_DAPLA_TEAM_API_URL
const USERS_URL = `${DAPLA_TEAM_API_URL}/users`
const TEAMS_URL = `${DAPLA_TEAM_API_URL}/teams`

export interface TeamOverviewData {
  [key: string]: TeamsData // myTeams, allTeams
}

interface TeamsData {
  teams: Team[]
}

export interface Team {
  uniform_name: string
  division_name: string
  display_name: string
  section_name: string
  section_code: string
  manager: TeamManager
  users: User[]
  groups: Group[]
  // eslint-disable-next-line
  _embedded?: any
}

interface TeamManager {
  display_name: string
  principal_name: string
}

interface User {
  display_name: string
  principal_name: string
}

interface Group {
  uniform_name: string
  display_name: string
  users: User[]
}

const fetchAllTeams = async (accessToken: string): Promise<TeamsData> => {
  const teamsUrl = new URL(`${TEAMS_URL}`)
  const embeds = ['users', 'groups.users']

  const selects = [
    'display_name',
    'uniform_name',
    'section_name',
    'users.principal_name',
    'groups.users.display_name',
    'groups.users.principal_name',
  ]

  teamsUrl.searchParams.set('embed', embeds.join(','))
  teamsUrl.searchParams.append('select', selects.join(','))

  try {
    const teams = await fetchAPIData(teamsUrl.toString(), accessToken)
    if (!teams) throw new ApiError(500, 'No json data returned')
    if (!teams._embedded || !teams._embedded.teams) return {} as TeamsData

    const flattedTeams = flattenEmbedded({ ...teams })
    flattedTeams.teams.forEach((team: Team, teamIndex: number) => {
      if (!team.groups) flattedTeams.teams[teamIndex].groups = []

      team.groups.forEach((group: Group, groupIndex: number) => {
        if (!group.users) {
          flattedTeams.teams[teamIndex].groups[groupIndex].users = []
        }
      })
      if (!team.users) flattedTeams.teams[teamIndex].users = []
    })

    const flattedTeamsWithManager = flattedTeams.teams.map((team: Team) => {
      const managers = team.groups.find((group) => group.uniform_name === `${team.uniform_name}-managers`)
      return {
        ...team,
        manager:
          managers && managers.users && managers.users.length > 0
            ? { display_name: managers.users[0].display_name, principal_name: managers.users[0].principal_name }
            : { display_name: 'Mangler manager', principal_name: 'ManglerManager@ssb.no' },
      }
    })

    return { teams: flattedTeamsWithManager }
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to fetch all teams:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('Failed to fetch all teams:', apiError)
      throw apiError
    }
  }
}

const fetchTeamsForPrincipalName = async (accessToken: string, principalName: string): Promise<TeamsData> => {
  const usersUrl = new URL(`${USERS_URL}/${principalName}`)
  const embeds = ['teams', 'teams.users', 'teams.groups.users']

  const selects = [
    'display_name',
    'principal_name',
    'teams.display_name',
    'teams.uniform_name',
    'teams.section_name',
    'teams.users.principal_name',
    'teams.groups.users.display_name',
    'teams.groups.users.principal_name',
  ]

  usersUrl.searchParams.set('embed', embeds.join(','))
  usersUrl.searchParams.append('select', selects.join(','))

  try {
    const teams = await fetchAPIData(usersUrl.toString(), accessToken)

    if (!teams) throw new ApiError(500, 'No json data returned')
    if (!teams._embedded || !teams._embedded.teams) return {} as TeamsData

    const flattedTeams = flattenEmbedded({ ...teams })
    flattedTeams.teams.forEach((team: Team) => {
      if (!team.groups) flattedTeams.teams.groups = []
      if (!team.users) flattedTeams.teams.users = []
    })

    const flattedTeamsWithManager = flattedTeams.teams.map((team: Team) => {
      const managers = team.groups.find((group) => group.uniform_name === `${team.uniform_name}-managers`)
      return {
        ...team,
        manager:
          managers && managers.users && managers.users.length > 0
            ? { display_name: managers.users[0].display_name, principal_name: managers.users[0].principal_name }
            : { display_name: 'Mangler ansvarlig', principal_name: 'ManglerAnsvarlig@ssb.no' },
      }
    })

    return { teams: flattedTeamsWithManager }
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to fetch teams for PrincipalName:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('Failed to fetch teams for PrincipalName:', apiError)
      throw apiError
    }
  }
}

export const fetchTeamOverviewData = async (principalName: string): Promise<TeamOverviewData> => {
  const accessToken = localStorage.getItem('access_token')
  if (!accessToken) {
    console.error('No access token available')
    const apiError = new ApiError(401, 'No access token available')
    console.error('Failed to fetch team members data:', apiError)
    throw apiError
  }

  try {
    const [myTeams, allTeams] = await Promise.all([
      fetchTeamsForPrincipalName(accessToken, principalName),
      fetchAllTeams(accessToken),
    ])

    return { myTeams: myTeams, allTeams: allTeams } as TeamOverviewData
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to fetch team overview data:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred while fetching team members data')
      console.error('Failed to fetch team overview data:', apiError)
      throw apiError
    }
  }
}
