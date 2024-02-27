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
  // eslint-disable-next-line
  _embedded?: any
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
      const errorMessage = (await response.text()) || 'An error occurred'
      const { detail, status } = JSON.parse(errorMessage)
      throw new ApiError(status, detail)
    }

    const jsonData = await response.json()

    if (!jsonData) throw new ApiError(500, 'No json data returned')
    if (!jsonData._embedded || !jsonData._embedded.managed_users) return [] // return an empty list if the user does not have any managed_users

    return jsonData._embedded.managed_users
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to fetch managed users:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('Failed to fetch managed users:', apiError)
      throw apiError
    }
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
          const { detail, status } = JSON.parse(errorMessage)
          throw new ApiError(status, detail)
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
    if (error instanceof ApiError) {
      console.error('Failed to fetch managed users:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('Failed to fetch managed users:', apiError)
      throw apiError
    }
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
      const { detail, status } = JSON.parse(errorMessage)
      throw new ApiError(status, detail)
    }

    const jsonData = await response.json()

    if (!jsonData) throw new ApiError(500, 'No json data returned')
    if (!jsonData._embedded || !jsonData._embedded.users) throw new ApiError(500, 'Did not receive users data')

    const transformedData = jsonData._embedded.users.map((user: User) => {
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
    if (error instanceof ApiError) {
      console.error('Failed to fetch all users:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('Failed to fetch all users:', apiError)
      throw apiError
    }
  }
}

export const fetchAllTeamMembersData = async (principalName: string): Promise<TeamMembersData | ApiError> => {
  const accessToken = localStorage.getItem('access_token')
  if (!accessToken) {
    console.error('No access token available')
    const apiError = new ApiError(401, 'No access token available')
    console.error('Failed to fetch team members data:', apiError)
    throw apiError
  }

  try {
    const [myUsers, allUsers] = await Promise.all([
      fetchManagersManagedUsers(accessToken, principalName),
      fetchAllUsers(accessToken),
    ])

    return { myUsers: myUsers, allUsers: allUsers } as TeamMembersData
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to fetch team members data:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred while fetching team members data')
      console.error('Failed to fetch team members data:', apiError)
      throw apiError
    }
  }
}
