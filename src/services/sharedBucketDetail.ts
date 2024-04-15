import { ApiError, fetchAPIData } from '../utils/services'
import { flattenEmbedded, DAPLA_TEAM_API_URL } from '../utils/utils'

const TEAMS_URL = `${DAPLA_TEAM_API_URL}/teams`

export interface SharedBucketDetail {
  [key: string]: Team | SharedBucket | Team[]
}

export interface SharedBucket {
  short_name: string
  bucket_name: string
  groups: Group[]
}

export interface Team {
  uniform_name: string
  display_name?: string
  section_name?: string
}

export interface User {
  display_name: string
  principal_name: string
  section_name: string
}

export interface Group {
  uniform_name: string
  users?: User[]
}

const fetchTeamDetail = async (teamId: string): Promise<Team> => {
  const teamUrl = new URL(`${TEAMS_URL}/${teamId}`, window.location.origin)

  const selects = ['uniform_name', 'section_name', 'groups.uniform_name']

  teamUrl.searchParams.set('selects', selects.join(','))

  try {
    const teamDetail = await fetchAPIData(teamUrl.toString())
    if (!teamDetail) throw new ApiError(500, 'No json data returned')

    const flattenedTeamDetail = flattenEmbedded({ ...teamDetail })

    return flattenedTeamDetail
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to fetch team detail:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('Failed to fetch team detail:', apiError)
      throw apiError
    }
  }
}

export const fetchSharedBucketDetailData = async (teamId: string, shortName: string): Promise<SharedBucket> => {
  const sharedBucketUrl = new URL(`${TEAMS_URL}/${teamId}/shared/buckets/${shortName}`, window.location.origin)

  const embeds = ['groups', 'groups.users']
  const selects = [
    'groups.uniform_name',
    'groups.users.principal_name',
    'groups.users.display_name',
    'groups.users.section_name',
  ]

  sharedBucketUrl.searchParams.set('embed', embeds.join(','))
  sharedBucketUrl.searchParams.set('selects', selects.join(','))

  try {
    const sharedBucket = await fetchAPIData(sharedBucketUrl.toString())
    if (!sharedBucket) throw new ApiError(500, 'No json data returned')

    const flattenedSharedBuckets = flattenEmbedded({ ...sharedBucket })
    if (!flattenedSharedBuckets.teams) flattenedSharedBuckets.teams = []

    return flattenedSharedBuckets
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to fetch shared buckets detail:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('Failed to fetch shared buckets detail:', apiError)
      throw apiError
    }
  }
}

export const fetchAllTeams = async (): Promise<Team[]> => {
  const teamsUrl = new URL(`${TEAMS_URL}`, window.location.origin)

  const selects = ['uniform_name']

  teamsUrl.searchParams.append('select', selects.join(','))

  try {
    const teams = await fetchAPIData(teamsUrl.toString())
    if (!teams) throw new ApiError(500, 'No json data returned')
    if (!teams._embedded?.teams) return {} as Team[]

    const flattedTeams = flattenEmbedded({ ...teams })
    return flattedTeams.teams
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

export const getSharedBucketDetailData = async (teamId: string, shortName: string): Promise<SharedBucketDetail> => {
  try {
    const [team, sharedBucket, allTeams] = await Promise.all([
      fetchTeamDetail(teamId),
      fetchSharedBucketDetailData(teamId, shortName),
      fetchAllTeams(),
    ])

    return { team, sharedBucket, allTeams }
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to fetch shared bucket detail:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred while fetching shared bucket data')
      console.error('Failed to fetch shared bucket detail:', apiError)
      throw apiError
    }
  }
}
