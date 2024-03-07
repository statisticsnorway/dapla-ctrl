import { useEffect, useState } from 'react'
import { useNavigate, Outlet } from 'react-router-dom'
import { getUserProfile } from '../services/userProfile'
import { fetchUserInformationFromAuthToken } from '../utils/services'

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
        localStorage.setItem('userProfile', JSON.stringify(await getUserProfile(userProfileData.email)))
        setIsAuthenticated(true)
      } catch (error) {
        console.error('Error occurred when updating userProfile data')
      }
    }
    setData()
  }, [from, navigate])

  return isAuthenticated ? <Outlet /> : null
}

export default ProtectedRoute
