import { DAPLA_TEAM_API_URL, flattenEmbedded } from '../utils/utils'

// eslint-disable-next-line
export const fetchAPIData = async (url: string): Promise<any> => {
  const response = await fetch(url)

  if (!response.ok) {
    const errorMessage = (await response.text()) || 'An error occurred'
    const { detail, status } = JSON.parse(errorMessage)
    throw new ApiError(status, detail)
  }

  return response.json()
}

export interface TokenData {
  name: string
  given_name: string
  family_name: string
  email: string
}

export const fetchUserInformationFromAuthToken = async (): Promise<TokenData> => {
  const response = await fetch('/localApi/fetch-token')

  const tokenData = await response.json()
  const jwt = JSON.parse(atob(tokenData.token.split('.')[1]))
  return { ...jwt } as TokenData
}

interface Group {
  uniform_name: string
  users: {
    principal_name: string
  }[]
}

const GROUPS_URL = `${DAPLA_TEAM_API_URL}/groups`

const fetchGroupMembership = async (groupUniformName: string): Promise<Group> => {
  const groupsUrl = new URL(`${GROUPS_URL}/${groupUniformName}/users`, window.location.origin)
  const embeds = ['users']
  const selects = ['uniform_name', 'users.principal_name']

  groupsUrl.searchParams.set('embed', embeds.join(','))
  groupsUrl.searchParams.append('select', selects.join(','))

  try {
    const groupDetail = await fetchAPIData(groupsUrl.toString())

    return flattenEmbedded(groupDetail)
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to fetch group membership:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('Failed to fetch group membership:', apiError)
      throw apiError
    }
  }
}

export const isDaplaAdmin = async (userPrincipalName: string): Promise<boolean> => {
  const adminGroups = import.meta.env.DAPLA_CTRL_ADMIN_GROUPS ?? ''
  console.log(`import.meta.env`, import.meta.env)
  const daplaAdminGroupsSeperated: string[] = adminGroups.split(',')
  if (daplaAdminGroupsSeperated.length === 0) return false

  try {
    const adminGroupUsers = await Promise.all(
      daplaAdminGroupsSeperated.map((groupUniformName: string) => fetchGroupMembership(groupUniformName))
    )

    return adminGroupUsers.some(
      (group) =>
        group.users && group.users.length > 0 && group.users.some((user) => user.principal_name === userPrincipalName)
    )
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to fetch dapla admins:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('Failed to fetch dapla admins:', apiError)
      throw apiError
    }
  }
}
export class ApiError extends Error {
  public code: number

  constructor(code: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.code = code
  }
}
