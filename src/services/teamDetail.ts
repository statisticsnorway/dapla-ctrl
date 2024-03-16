import { ApiError, fetchAPIData } from '../utils/services'
import { flattenEmbedded, DAPLA_TEAM_API_URL } from '../utils/utils'

const TEAMS_URL = `${DAPLA_TEAM_API_URL}/teams`
const GROUPS_URL = `${DAPLA_TEAM_API_URL}/groups`

export interface TeamDetailData {
  [key: string]: Team | SharedBuckets // teamUsers, sharedBuckets
}

export interface Team {
  uniform_name: string
  division_name?: string
  display_name: string
  section_name: string
  section_code?: string
  section_manager: User
  managers?: TeamManager[]
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
  groups?: Group[]
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
  metrics?: Metrics
}

export interface Metrics {
  teams_count?: number | string
  groups_count?: number | string
  users_count?: number | string
}

export interface JobResponse {
  status: string
  detail?: string
}

export const fetchTeamInfo = async (teamId: string): Promise<Team | ApiError> => {
  const teamsUrl = new URL(`${TEAMS_URL}/${teamId}`, window.location.origin)
  const embeds = ['users', 'users.groups', 'managers', 'groups']
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
    'groups.uniform_name',
  ]

  teamsUrl.searchParams.set('embed', embeds.join(','))
  teamsUrl.searchParams.append('select', selects.join(','))

  try {
    const teamDetailData = await fetchAPIData(teamsUrl.toString())
    const flattendTeams = flattenEmbedded(teamDetailData)
    if (!flattendTeams) return {} as Team
    if (!flattendTeams.users) flattendTeams.users = []
    flattendTeams.users.forEach((user: User) => {
      if (!user.section_name || user.section_name === '') user.section_name = 'Mangler seksjon'
    })
    if (!flattendTeams.groups) flattendTeams.groups = []
    if (!flattendTeams.managers || flattendTeams.managers.length === 0) {
      flattendTeams.manager = {
        display_name: 'Ikke funnet',
        principal_name: 'Ikke funnet',
        section_name: 'Ikke funnet',
      }
    }

    if (flattendTeams._links && flattendTeams._links.section_manager) {
      const sectionManagerUrl = new URL(flattendTeams._links.section_manager.href)
      const path = sectionManagerUrl.pathname
      flattendTeams.section_manager = await fetchTeamSectionManager(`${DAPLA_TEAM_API_URL}${path}`)
    } else {
      flattendTeams.section_manager = { display_name: 'Seksjonsleder ikke funnet', principal_name: 'Ikkefunnet@ssb.no' }
    }

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

export const fetchSharedBuckets = async (teamId: string): Promise<SharedBuckets | ApiError> => {
  const sharedBucketsUrl = new URL(`${TEAMS_URL}/${teamId}/shared/buckets`, window.location.origin)

  const embeds = ['metrics']
  const selects = ['short_name', 'bucket_name', 'metrics.teams_count', 'metrics.groups_count', 'metrics.users_count']

  sharedBucketsUrl.searchParams.set('embed', embeds.join(','))
  sharedBucketsUrl.searchParams.append('select', selects.join(','))

  try {
    const sharedBuckets = await fetchAPIData(sharedBucketsUrl.toString())
    if (!sharedBuckets) throw new ApiError(500, 'No json data returned')
    if (!sharedBuckets._embedded) return {} as SharedBuckets

    const flattenedSharedBuckets = flattenEmbedded({ ...sharedBuckets })
    flattenedSharedBuckets.items.forEach((item: SharedBucket) => {
      if (!item.metrics) {
        item.metrics = {
          teams_count: 'Ingen data',
          groups_count: 'Ingen data',
          users_count: 'Ingen data',
        }
      } else {
        item.metrics = (item.metrics as Metrics[])[0]
      }
    })

    return flattenedSharedBuckets
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to fetch shared buckets:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('Failed to fetch shared buckets:', apiError)
      throw apiError
    }
  }
}

const fetchTeamSectionManager = async (url: string) => {
  try {
    const response = await fetchAPIData(url)

    return response
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to fetch team section manager:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('Failed to fetch team section manager:', apiError)
      throw apiError
    }
  }
}

export const getTeamDetail = async (teamId: string): Promise<TeamDetailData> => {
  try {
    const [teamInfo, sharedBuckets] = await Promise.all([fetchTeamInfo(teamId), fetchSharedBuckets(teamId)])

    return { team: teamInfo as Team, sharedBuckets } as TeamDetailData
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to fetch data for teamDetail page:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('Failed to fetch data for teamDetail page:', apiError)
      throw apiError
    }
  }
}

export const addUserToGroups = async (groupIds: string[], userPrincipalName: string): Promise<JobResponse[]> => {
  try {
    const jobResponses = await Promise.all(groupIds.map((groupId) => addUserToGroup(groupId, userPrincipalName)))
    return jobResponses
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to add user to groups: ', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('Failed to add user to groups: ', apiError)
      throw apiError
    }
  }
}

const addUserToGroup = async (groupId: string, userPrincipalName: string): Promise<JobResponse> => {
  const groupsUrl = `${GROUPS_URL}/${groupId}/users`
  try {
    const response = await fetch(groupsUrl, {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify({
        users: [userPrincipalName],
      }),
    })

    if (!response.ok) {
      const errorMessage = (await response.text()) || 'An error occurred'
      const { detail, status } = JSON.parse(errorMessage)
      throw new ApiError(status, detail)
    }

    const responseJson = await response.json()
    const flattendResponse = { ...responseJson._embedded.results[0] }

    return flattendResponse
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to add user to group: ', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('Failed to add user to group: ', apiError)
      throw apiError
    }
  }
}
