import React, { useEffect, useState } from 'react'
import NotFound from '../pages/NotFound/NotFound.tsx'
import { User } from '../services/userProfile'
import { isAuthorizedToCreateTeam } from '../services/createTeam'
import { Effect } from 'effect'
import { isDaplaAdmin } from '../utils/services'

export interface Props {
  component: React.ReactElement
}

const ProtectedAuthorizedUserRoute = ({ component }: Props) => {
  const [isAuthorized, setIsAuthorized] = useState(false)
  useEffect(() => {
    const userProfileItem = localStorage.getItem('userProfile')
    if (!userProfileItem) return

    const user = JSON.parse(userProfileItem) as User
    if (!user) return

    Effect.promise(() => isDaplaAdmin(user.principal_name))
      .pipe(Effect.runPromise)
      .then((isDaplaAdmin: boolean) => setIsAuthorized(isAuthorizedToCreateTeam(isDaplaAdmin, user.job_title)))
  }, [])

  return isAuthorized ? component : <NotFound />
}

export default ProtectedAuthorizedUserRoute
