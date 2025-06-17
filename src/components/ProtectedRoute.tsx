import { useEffect, useState } from 'react'
import { useNavigate, Outlet } from 'react-router-dom'
import { fetchUserInformationFromAuthToken } from '../utils/services'
import { Effect, Option as O, Schema } from 'effect'
import { ParseError } from 'effect/ParseResult'
import { HttpClientError } from '@effect/platform/HttpClientError'
import { customLogger } from '../utils/logger.ts'
import { useUserProfileStore } from '../services/store.ts'
import { UserProfile } from '../@types/user.ts'
import { getUserProfileE } from '../services/userProfile'

const ProtectedRoute = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const setLoggedInUser = useUserProfileStore((state) => state.setLoggedInUser)
  const navigate = useNavigate()
  const from = location.pathname

  useEffect(() => {
    const fetchAndStoreUserProfile = (): Effect.Effect<void, HttpClientError | ParseError | Error> =>
      Effect.gen(function* () {
        const userProfileData = yield* Effect.promise(fetchUserInformationFromAuthToken)
        const userProfile: UserProfile = yield* getUserProfileE(userProfileData.email)
        yield* Effect.sync(() => {
          localStorage.setItem('userProfile', JSON.stringify(userProfile))
          setLoggedInUser(userProfile)
          setIsAuthenticated(true)
        })
      }).pipe(Effect.provide(customLogger))

    const cachedUserProfile: O.Option<UserProfile> = O.fromNullable(localStorage.getItem('userProfile')).pipe(
      O.flatMap(Schema.decodeUnknownOption(UserProfile))
    )

    O.match(cachedUserProfile, {
      onNone: () => fetchAndStoreUserProfile(),
      onSome: (userProfile) =>
        Effect.sync(() => {
          setIsAuthenticated(true)
          setLoggedInUser(userProfile)
        }),
    }).pipe(Effect.runPromise)
  }, [from, navigate, setLoggedInUser])

  return isAuthenticated ? <Outlet /> : null
}

export default ProtectedRoute
