import { useEffect, useState } from 'react'
import { useNavigate, Outlet } from 'react-router-dom'
import { getUserProfile } from '../services/userProfile'
import { fetchUserInformationFromAuthToken } from '../utils/services'
import { Effect } from 'effect'
import { customLogger } from '../utils/logger.ts'

const ProtectedRoute = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const navigate = useNavigate()
  const from = location.pathname

  useEffect(() => {
    const setData = async () => {
      try {
        if (localStorage.getItem('userProfile') !== null) {
          setIsAuthenticated(true)
          return
        }

        const userProfileData = await fetchUserInformationFromAuthToken()
        const userProfile = JSON.stringify(await getUserProfile(userProfileData.email))
        Effect.logInfo(`UserProfile set in localStorage: ${userProfile}`).pipe(
          Effect.provide(customLogger),
          Effect.runSync
        )
        localStorage.setItem('userProfile', userProfile)
        setIsAuthenticated(true)
      } catch (error) {
        console.error('Error occurred when updating userProfile data')
        console.error(error)
      }
    }
    setData()
  }, [from, navigate])

  return isAuthenticated ? <Outlet /> : null
}

export default ProtectedRoute
