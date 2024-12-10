import { useEffect, useState } from 'react'
import NotFound from '../pages/NotFound/NotFound.tsx'
import { User } from '../services/userProfile'
import { isAuthorizedToCreateTeam } from '../services/createTeam'
import { Effect, Option as O } from 'effect'
import { isDaplaAdmin } from '../utils/services'
import { option } from '../utils/utils'
import { Skeleton } from '@mui/material'
import PageLayout from './PageLayout/PageLayout.tsx'
import { Outlet } from 'react-router-dom'
import { useUserProfileStore } from '../services/store.ts'

const ProtectedAuthorizedUserRoute = () => {
  // O.none() here represents the loading state
  const [oIsAuthorized, setIsAuthorized] = useState<O.Option<boolean>>(O.none())
  const maybeUser: O.Option<User> = useUserProfileStore((state) => state.loggedInUser)

  useEffect(() => {
    Effect.gen(function* () {
      const user: User = yield* O.match(maybeUser, {
        onNone: () => Effect.fail(new Error('User not logged in!')),
        onSome: (user) => Effect.succeed(user),
      })

      const daplaAdmin: boolean = yield* Effect.promise(() => isDaplaAdmin(user.principal_name))

      yield* Effect.sync(() => setIsAuthorized(O.some(isAuthorizedToCreateTeam(daplaAdmin, user.job_title))))
   }).pipe(Effect.runPromise)
  })

  return option(
    oIsAuthorized,
    () => (
      <PageLayout
        title='Opprett Team'
        content={<Skeleton variant='rectangular' animation='wave' width={800} height={600} />}
      />
    ),
    (isAuthorized) => (isAuthorized ? <Outlet /> : <NotFound />)
  )
}

export default ProtectedAuthorizedUserRoute
