import { ApiError, fetchAPIData } from '../utils/services'
import { flattenEmbedded } from '../utils/utils'

const DAPLA_TEAM_API_URL = import.meta.env.VITE_DAPLA_TEAM_API_URL
const TEAMS_URL = `${DAPLA_TEAM_API_URL}/teams`

export interface SharedBucketDetail {
  [key: string]: Team | SharedBucket
}

export interface SharedBucket {
  short_name: string
  bucket_name: string
  teams: Team[]
}

export interface Team {
  uniform_name: string
  display_name?: string
  section_name: string
}

const fetchTeamDetail = async (teamId: string): Promise<Team> => {
  const teamUrl = new URL(`${TEAMS_URL}/${teamId}`)

  const selects = ['uniform_name', 'section_name']

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
  const sharedBucketUrl = new URL(`${TEAMS_URL}/${teamId}/shared/buckets/${shortName}`)

  const embeds = ['teams']
  const selects = ['teams.uniform_name', 'teams.display_name', 'teams.section_name']

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

export const getSharedBucketDetailData = async (teamId: string, shortName: string): Promise<SharedBucketDetail> => {
  try {
    const [team, sharedBucket] = await Promise.all([
      fetchTeamDetail(teamId),
      fetchSharedBucketDetailData(teamId, shortName),
    ])

    return { team, sharedBucket }
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
