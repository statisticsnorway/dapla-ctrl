import { ApiError, fetchAPIData } from '../utils/services'
import { DAPLA_TEAM_API_URL } from '../utils/utils'

const USERS_URL = `${DAPLA_TEAM_API_URL}/users`

export interface TeamMembersData {
  [key: string]: UsersData // myUsers, allUsers
}

export interface UsersData {
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

const fetchManagedUsers = async (principalName: string): Promise<User[]> => {
  const usersUrl = new URL(`${USERS_URL}/${principalName}`, window.location.origin)
  const embeds = ['managed_users']
  const selects = ['managed_users.principal_name']

  usersUrl.searchParams.set('embed', embeds.join(','))
  usersUrl.searchParams.append('select', selects.join(','))

  try {
    const managedUsersData = await fetchAPIData(usersUrl.toString())

    if (!managedUsersData) throw new ApiError(500, 'No json data returned')
    if (!managedUsersData._embedded || !managedUsersData._embedded.managed_users) return [] // return an empty list if the user does not have any managed_users

    return managedUsersData._embedded.managed_users
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

export const fetchManagedUsersManagers = async (principalName: string): Promise<UsersData> => {
  try {
    const users = await fetchManagedUsers(principalName)

    const prepUsers = await Promise.all(
      users.map(async (user): Promise<User> => {
        const usersUrl = new URL(`${USERS_URL}/${user.principal_name}`, window.location.origin)
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

        const managedUsersData = await fetchAPIData(usersUrl.toString())

        const prepData = {
          ...managedUsersData,
          ...managedUsersData._embedded,
        }
        delete prepData._embedded

        return prepData
      })
    )

    return { users: prepUsers }
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

export const fetchAllUsers = async (): Promise<UsersData> => {
  const usersUrl = new URL(`${USERS_URL}`, window.location.origin)
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
    const allUsersData = await fetchAPIData(usersUrl.toString())

    if (!allUsersData) throw new ApiError(500, 'No json data returned')
    if (!allUsersData._embedded || !allUsersData._embedded.users) throw new ApiError(500, 'Did not receive users data')

    const prepData = allUsersData._embedded.users.map((user: User) => {
      const prepUserData = {
        ...user,
        ...user._embedded,
      }
      delete prepUserData._embedded
      return prepUserData
    })
    delete prepData._embedded

    return { users: prepData }
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
  try {
    const [myUsers, allUsers] = await Promise.all([fetchManagedUsersManagers(principalName), fetchAllUsers()])

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

export const fetchUserSearchData = async (): Promise<User[]> => {
  const usersUrl = new URL(`${USERS_URL}`, window.location.origin)
  const selects = ['display_name', 'principal_name', 'section_name']

  usersUrl.searchParams.append('select', selects.join(','))

  try {
    const allUsersData = await fetchAPIData(usersUrl.toString())

    if (!allUsersData) throw new ApiError(500, 'No json data returned')
    if (!allUsersData._embedded || !allUsersData._embedded.users) throw new ApiError(500, 'Did not receive users data')

    const prepData = allUsersData._embedded.users.map((user: User) => {
      const prepUserData = {
        ...user,
        ...user._embedded,
      }
      delete prepUserData._embedded
      return prepUserData
    })
    delete prepData._embedded

    return prepData
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to fetch user search data:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('Failed to fetch user search data:', apiError)
      throw apiError
    }
  }
}
