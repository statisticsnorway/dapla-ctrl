import { ApiError, fetchAPIData } from '../utils/services'
import { flattenEmbedded, DAPLA_TEAM_API_URL } from '../utils/utils'

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
  managers: TeamManager[]
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

const fetchAllTeams = async (): Promise<TeamsData> => {
  const teamsUrl = new URL(`${TEAMS_URL}`, window.location.origin)
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
    const teams = await fetchAPIData(teamsUrl.toString())
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
      const managersGroup = team.groups.find((group) => group.uniform_name === `${team.uniform_name}-managers`)

      const managers =
        managersGroup && managersGroup.users && managersGroup.users.length > 0
          ? managersGroup.users.map((manager) => ({
              display_name: manager.display_name,
              principal_name: manager.principal_name,
            }))
          : [
              {
                display_name: 'Mangler managers',
                principal_name: 'ManglerManagers@ssb.no',
              },
            ]

      return {
        ...team,
        managers: managers,
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

const fetchTeamsForPrincipalName = async (principalName: string): Promise<TeamsData> => {
  const usersUrl = new URL(`${USERS_URL}/${principalName}`, window.location.origin)
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
    const teams = await fetchAPIData(usersUrl.toString())

    if (!teams) throw new ApiError(500, 'No json data returned')
    if (!teams._embedded || !teams._embedded.teams) return {} as TeamsData

    const flattedTeams = flattenEmbedded({ ...teams })
    flattedTeams.teams.forEach((team: Team) => {
      if (!team.groups) flattedTeams.teams.groups = []
      if (!team.users) flattedTeams.teams.users = []
    })

    const flattedTeamsWithManager = flattedTeams.teams.map((team: Team) => {
      const managersGroup = team.groups.find((group) => group.uniform_name === `${team.uniform_name}-managers`)

      const managers =
        managersGroup && managersGroup.users && managersGroup.users.length > 0
          ? managersGroup.users.map((manager) => ({
              display_name: manager.display_name,
              principal_name: manager.principal_name,
            }))
          : [
              {
                display_name: 'Mangler manager',
                principal_name: 'ManglerManagers@ssb.no',
              },
            ]

      return {
        ...team,
        managers: managers,
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
  try {
    const [myTeams, allTeams] = await Promise.all([fetchTeamsForPrincipalName(principalName), fetchAllTeams()])

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
