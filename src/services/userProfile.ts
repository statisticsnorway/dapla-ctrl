import { ApiError, fetchAPIData } from '../utils/services'
import { flattenEmbedded, DAPLA_TEAM_API_URL } from '../utils/utils'
import { ConversionError } from '../@types/error'
import { UserData, UserProfile, UserPhoto } from '../@types/user'

import { Effect, ParseResult } from 'effect'
import { FetchHttpClient, HttpClient, HttpClientRequest, HttpClientResponse, HttpClientError } from '@effect/platform'

const USERS_URL = `${DAPLA_TEAM_API_URL}/users`

export interface UserProfileTeamData {
  [key: string]: TeamsData
}

export interface TeamsData {
  user: User
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

export interface User {
  display_name: string
  principal_name: string
  section_name: string
  section_code: string
  job_title: string
  azure_ad_id?: string
  first_name?: string
  last_name?: string
  email?: string
}

interface Group {
  uniform_name: string
  display_name: string
  users: User[]
}

export const getUserSectionCode = (
  principalName: string
): Effect.Effect<number, Error | HttpClientError.HttpClientError> =>
  HttpClient.HttpClient.pipe(
    Effect.flatMap((client) =>
      HttpClientRequest.get(new URL(`${USERS_URL}/${principalName}`, window.location.origin)).pipe(
        HttpClientRequest.appendUrlParam('select', 'section_code'),
        client.execute,
        Effect.flatMap((res) => res.json)
      )
    ),
    Effect.flatMap((jsonResponse: unknown) =>
      Effect.try({
        // We need a type assertion here because `jsonResponse` is of type unknown
        try: () => (jsonResponse as { section_code: number }).section_code,
        catch: (error) => new Error(`Failed to get section_code: ${error}`),
      })
    )
  ).pipe(Effect.scoped, Effect.provide(FetchHttpClient.layer))

export const getUserData = (
  principalName: string
): Effect.Effect<UserData, HttpClientError.HttpClientError | ParseResult.ParseError> =>
  HttpClient.HttpClient.pipe(
    Effect.flatMap((client) =>
      HttpClientRequest.get(new URL(`${USERS_URL}/${principalName}`, window.location.origin)).pipe(client.execute)
    ),
    Effect.flatMap(HttpClientResponse.schemaBodyJson(UserData)),
    Effect.scoped,
    Effect.provide(FetchHttpClient.layer)
  )

// Convert base64 string to Blob URL while handling potential errors
const base64ToBlobUrl = (base64Image: string): Effect.Effect<string, ConversionError> =>
  Effect.try({
    try: () => {
      const byteArray = new Uint8Array(Array.from(atob(base64Image), (c) => c.charCodeAt(0)))
      const blob = new Blob([byteArray], { type: 'image/png' })
      return URL.createObjectURL(blob)
    },
    catch: (unknown) =>
      new ConversionError(`Failed to convert base64 avatar photo to Blob URL: ${(unknown as Error).message}`),
  })

// Fetch a users data and photo, then combine them to the UserProfile type
export const getUserProfileE = (
  principalName: string
): Effect.Effect<UserProfile, HttpClientError.HttpClientError | ParseResult.ParseError | Error> =>
  Effect.gen(function* () {
    const userData = yield* getUserData(principalName)

    const base64Image: UserPhoto = yield* HttpClient.HttpClient.pipe(
      Effect.flatMap((client) => HttpClientRequest.get(`/localApi/photo/${principalName}`).pipe(client.execute)),
      Effect.flatMap(HttpClientResponse.schemaBodyJson(UserPhoto)),
      Effect.scoped,
      Effect.provide(FetchHttpClient.layer)
    )

    const photoBlobUrl: string = yield* base64ToBlobUrl(base64Image.photo)

    return { ...userData, photo: photoBlobUrl }
  })

// TODO: Remove this function once its last use site in `getUserProfileTeamData` is
// rewritten
export const getUserProfile = async (principalName: string): Promise<User | ApiError> => {
  const usersUrl = new URL(`${USERS_URL}/${principalName}`, window.location.origin)
  const selects = [
    'principal_name',
    'display_name',
    'first_name',
    'last_name',
    'section_name',
    'section_code',
    'division_name',
    'phone',
    'job_title',
  ]

  usersUrl.searchParams.set('select', selects.join(','))

  try {
    const [userData, userPhoto] = await Promise.all([fetchAPIData(usersUrl.toString()), fetchPhoto(principalName)])

    userData.photo = userPhoto

    return { ...userData }
  } catch (error) {
    if (error instanceof ApiError) {
      console.error('Failed to fetch userProfile data:', error)
      throw error
    } else {
      const apiError = new ApiError(500, 'An unexpected error occurred')
      console.error('Failed to fetch userProfile data:', apiError)
      throw apiError
    }
  }
}

export const getUserProfileTeamData = async (principalName: string): Promise<TeamsData | ApiError> => {
  const usersUrl = new URL(`${USERS_URL}/${principalName}`, window.location.origin)
  const embeds = ['teams', 'teams.groups', 'teams.groups.users']

  const selects = [
    'display_name',
    'principal_name',
    'teams.section_name',
    'teams.display_name',
    'teams.uniform_name',
    'teams.groups.uniform_name',
    'teams.groups.users.principal_name',
    'teams.groups.users.display_name',
  ]

  usersUrl.searchParams.set('embed', embeds.join(','))
  usersUrl.searchParams.append('select', selects.join(','))

  try {
    const userProfileData = await fetchAPIData(usersUrl.toString())

    if (!userProfileData) throw new ApiError(500, 'No json data returned')
    if (!userProfileData._embedded || !userProfileData._embedded.teams) return {} as TeamsData

    const flattedTeams = flattenEmbedded({ ...userProfileData })
    flattedTeams.teams.forEach((team: Team, teamIndex: number) => {
      if (!team.groups) flattedTeams.teams.groups = []

      team.groups.forEach((group: Group, groupIndex: number) => {
        if (!group.users) flattedTeams.teams[teamIndex].groups[groupIndex].users = []
      })
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

    const getUser = await getUserProfile(principalName)

    return { teams: flattedTeamsWithManager, user: getUser as User }
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

const fetchPhoto = async (principalName: string) => {
  const response = await fetch(`/localApi/photo/${principalName}`)

  if (!response.ok) {
    throw new ApiError(500, 'could not fetch photo')
  }

  const data = await response.json()
  return data.photo
}
