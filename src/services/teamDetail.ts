import { ApiError, fetchAPIData } from '../utils/services'
import { flattenEmbedded } from '../utils/utils'

const DAPLA_TEAM_API_URL = import.meta.env.VITE_DAPLA_TEAM_API_URL
const TEAMS_URL = `${DAPLA_TEAM_API_URL}/teams`

export interface TeamDetailData {
  [key: string]: Team | SharedBuckets // teamUsers, sharedBuckets
}

export interface Team {
  uniform_name: string
  division_name?: string
  display_name: string
  section_name: string
  section_code?: string
  manager?: TeamManager
  users?: User[]
  groups?: Group[]
  // eslint-disable-next-line
  _embedded?: any
}

interface TeamManager {
  display_name: string
  principal_name: string
}

export interface User {
  display_name: string
  principal_name: string
  section_name: string
  groups: Group[]
}

interface Group {
  uniform_name: string
  display_name: string
}

export interface SharedBuckets {
  items: SharedBucket[]
  // eslint-disable-next-line
    _embedded?: any
}

export interface SharedBucket {
  short_name: string
  bucket_name: string
  metrics?: Metrics[]
}

interface Metrics {
  teams_count: number
  groups_count: number
  users_count: number
}

export const fetchTeamInfo = async (teamId: string, accessToken: string): Promise<Team | ApiError> => {
  const teamsUrl = new URL(`${TEAMS_URL}/${teamId}`)
  const embeds = ['users', 'users.groups', 'managers']
  const selects = [
    'uniform_name',
    'display_name',
    'section_name',
    'managers.principal_name',
    'managers.display_name',
    'managers.section_name',
    'users.principal_name',
    'users.display_name',
    'users.section_name',
    'users.groups.uniform_name',
  ]

  teamsUrl.searchParams.set('embed', embeds.join(','))
  teamsUrl.searchParams.append('select', selects.join(','))

  try {
    const teamDetailData = await fetchAPIData(teamsUrl.toString(), accessToken)
    const flattendTeams = flattenEmbedded(teamDetailData)
    if (!flattendTeams) return {} as Team
    if (!flattendTeams.users) flattendTeams.users = []
    if (!flattendTeams.managers || flattendTeams.managers.length === 0) {
      flattendTeams.manager = {
        display_name: 'Ikke funnet',
        principal_name: 'Ikke funnet',
        section_name: 'Ikke funnet',
      }
    } else {
      flattendTeams.manager = flattendTeams.managers[0]
    }
    delete flattendTeams.managers

    return flattendTeams
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to fetch teams:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('Failed to fetch teams:', apiError)
      throw apiError
    }
  }
}

export const fetchSharedBuckets = async (teamId: string, accessToken: string): Promise<SharedBuckets | ApiError> => {
  const sharedBucketsUrl = new URL(`${TEAMS_URL}/${teamId}/shared/buckets`)

  const embeds = ['metrics']
  const selects = ['short_name', 'bucket_name', 'metrics.teams_count', 'metrics.groups_count', 'metrics.users_count']

  sharedBucketsUrl.searchParams.set('embed', embeds.join(','))
  sharedBucketsUrl.searchParams.append('select', selects.join(','))

  try {
    const sharedBuckets = await fetchAPIData(sharedBucketsUrl.toString(), accessToken)
    if (!sharedBuckets) throw new ApiError(500, 'No json data returned')
    if (!sharedBuckets._embedded) return {} as SharedBuckets

    // TODO: Add fallback for teams with shared buckets that have no metrics
    const flattenedSharedBuckets = flattenEmbedded({ ...sharedBuckets })

    return flattenedSharedBuckets
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to fetch teams:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('Failed to fetch teams:', apiError)
      throw apiError
    }
  }
}

export const getTeamDetail = async (teamId: string): Promise<TeamDetailData> => {
  const accessToken = localStorage.getItem('access_token') as string

  try {
    const [teamInfo, sharedBuckets] = await Promise.all([
      fetchTeamInfo(teamId, accessToken),
      fetchSharedBuckets(teamId, accessToken),
    ])

    return { team: teamInfo as Team, sharedBuckets: sharedBuckets as SharedBuckets } as TeamDetailData
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to fetch data for teamDetail page:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('FFailed to fetch data for teamDetail page:', apiError)
      throw apiError
    }
  }
}
