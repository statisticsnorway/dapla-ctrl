import { useEffect, useState } from 'react'
import { useNavigate, Outlet } from 'react-router-dom'
import { getUserProfile } from '../services/userProfile'
import { fetchUserInformationFromAuthToken } from '../utils/services'
import { Cause, Effect, Option as O } from 'effect'
import { customLogger } from '../utils/logger.ts'
import { ApiError } from '../utils/services.ts'
import { useUserProfileStore } from '../services/store.ts'

import { User } from '../services/userProfile.ts'

const ProtectedRoute = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const setUser = useUserProfileStore((state) => state.setUser)
  const navigate = useNavigate()
  const from = location.pathname

  const fetchUserProfile = (): Effect.Effect<void, Cause.UnknownException | ApiError> =>
    Effect.gen(function* () {
      const userProfileData = yield* Effect.promise(fetchUserInformationFromAuthToken)
      const userProfile = yield* Effect.tryPromise(() => getUserProfile(userProfileData.email)).pipe(
        Effect.flatMap((x) => (x instanceof ApiError ? Effect.fail(x) : Effect.succeed(x)))
      )
      yield* Effect.sync(() => localStorage.setItem('userProfile', JSON.stringify(userProfile)))
      yield* Effect.sync(() => setUser(userProfile))
      yield* Effect.sync(() => setIsAuthenticated(true))
    }).pipe(Effect.provide(customLogger))

  useEffect(() => {
    const cachedUserProfile: O.Option<User> = O.fromNullable(localStorage.getItem('userProfile')).pipe(
      O.flatMap(O.liftThrowable(JSON.parse))
    )
    O.match(cachedUserProfile, {
      onNone: () => fetchUserProfile(),
      onSome: (userProfile) =>
        // invalidate cached user profile if 'job_title' field is missing
        userProfile.job_title
          ? Effect.zip(
              Effect.sync(() => setIsAuthenticated(true)),
              Effect.sync(() => setUser(userProfile))
            )
          : Effect.zipRight(
              Effect.logInfo("'job_title' field missing, invalidating UserProfile cache"),
              fetchUserProfile()
            ).pipe(Effect.provide(customLogger)),
    }).pipe(Effect.runPromise)
  }, [from, navigate, fetchUserProfile, setUser])

  return isAuthenticated ? <Outlet /> : null
}

export default ProtectedRoute
