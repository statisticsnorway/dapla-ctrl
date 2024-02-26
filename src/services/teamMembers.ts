import { ErrorResponse } from '../@types/error'
import { ApiError } from './ApiError'

const DAPLA_TEAM_API_URL = import.meta.env.VITE_DAPLA_TEAM_API_URL
const USERS_URL = `${DAPLA_TEAM_API_URL}/users`

export interface TeamMembersData {
  [key: string]: UsersData // myUsers, allUsers
}

interface UsersData {
  users: User[]
}

export interface User {
  principal_name: string
  display_name: string
  section_name: string
  section_manager: sectionManager[]
  teams: Team[]
  groups: Group[]
}

interface sectionManager {
  display_name: string
  principal_name: string
}

interface Team {
  uniform_name: string
}

interface Group {
  uniform_name: string
}

const fetchManagedUsers = async (accessToken: string, principalName: string): Promise<User[]> => {
  const usersUrl = new URL(`${USERS_URL}/${principalName}`)
  const embeds = ['managed_users']
  const selects = ['managed_users.principal_name']

  usersUrl.searchParams.set('embed', embeds.join(','))
  usersUrl.searchParams.append('select', selects.join(','))

  try {
    const response = await fetch(usersUrl.toString(), {
      method: 'GET',
      headers: {
        accept: '*/*',
        Authorization: `Bearer ${accessToken}`,
      },
    })

    if (!response.ok) {
      // TODO: Test that it actually works
      const errorMessage = (await response.text()) || 'An error occurred'
      throw new ApiError(response.status, errorMessage)
    }

    const jsonData = await response.json()

    if (!jsonData) throw new Error('No json data returned')
    if (!jsonData._embedded || !jsonData._embedded.managed_users) return [] // return an empty list if the user does not have any managed_users

    return jsonData._embedded.managed_users
  } catch (error) {
    throw error
  }
}

export const fetchManagersManagedUsers = async (accessToken: string, principalName: string): Promise<UsersData> => {
  try {
    const users = await fetchManagedUsers(accessToken, principalName)

    const formattedUsers = await Promise.all(
      users.map(async (user): Promise<User> => {
        const usersUrl = new URL(`${USERS_URL}/${user.principal_name}`)
        const embeds = ['teams', 'groups', 'section_manager']

        const selects = [
          'principal_name',
          'display_name',
          'section_name',
          'teams.uniform_name',
          'groups.uniform_name',
          'section_manager.display_name',
          'section_manager.principal_name',
        ]

        usersUrl.searchParams.set('embed', embeds.join(','))
        usersUrl.searchParams.append('select', selects.join(','))

        const response = await fetch(usersUrl.toString(), {
          method: 'GET',
          headers: {
            accept: '*/*',
            Authorization: `Bearer ${accessToken}`,
          },
        })

        if (!response.ok) {
          const errorMessage = (await response.text()) || 'An error occurred'
          throw new ApiError(response.status, errorMessage)
        }

        const jsonData = await response.json()
        const transformedData = {
          ...jsonData,
          ...jsonData._embedded,
        }
        delete transformedData._embedded

        return transformedData
      })
    )

    return { users: formattedUsers }
  } catch (error) {
    console.error('Failed to fetch managed users:', error)
    throw error
  }
}

export const fetchAllUsers = async (accessToken: string): Promise<UsersData> => {
  const usersUrl = new URL(`${USERS_URL}`)
  const embeds = ['section_manager', 'teams', 'groups']

  const selects = [
    'display_name',
    'principal_name',
    'section_name',
    'section_manager.display_name',
    'section_manager.principal_name',
    'teams.uniform_name',
    'groups.uniform_name',
  ]

  usersUrl.searchParams.set('embed', embeds.join(','))
  usersUrl.searchParams.append('select', selects.join(','))

  try {
    const response = await fetch(usersUrl.toString(), {
      method: 'GET',
      headers: {
        accept: '*/*',
        Authorization: `Bearer ${accessToken}`,
      },
    })

    if (!response.ok) {
      const errorMessage = (await response.text()) || 'An error occurred'
      throw new ApiError(response.status, errorMessage)
    }

    const jsonData = await response.json()

    if (!jsonData) throw new Error('No json data returned')
    if (!jsonData._embedded || !jsonData._embedded.users) throw new Error('Did not receive users data')

    const transformedData = jsonData._embedded.users.map((user: any) => {
      const userFormatted = {
        ...user,
        ...user._embedded,
      }
      delete userFormatted._embedded
      return userFormatted
    })
    delete transformedData._embedded

    return { users: transformedData }
  } catch (error) {
    console.error('Failed to fetch all users:', error)
    throw error
  }
}

export const fetchAllTeamMembersData = async (principalName: string): Promise<TeamMembersData | ErrorResponse> => {
  const accessToken = localStorage.getItem('access_token')
  if (!accessToken) {
    console.error('No access token available')
    throw new Error('No access token available')
  }

  try {
    const [myUsers, allUsers] = await Promise.all([
      fetchManagersManagedUsers(accessToken, principalName),
      fetchAllUsers(accessToken),
    ])

    return { myUsers: myUsers, allUsers: allUsers } as TeamMembersData
  } catch (error) {
    console.error('Failed to fetch all data for teamMembers:', error)
    throw error
  }
}
