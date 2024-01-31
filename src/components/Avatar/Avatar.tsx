import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { User } from '../../@types/user'
import styles from './avatar.module.scss'

export default function Avatar() {
  const [userProfileData, setUserProfileData] = useState<User>()
  const [imageSrc, setImageSrc] = useState<string>()
  const [fallbackInitials, setFallbackInitials] = useState<string>('??')
  const [encodedURI, setEncodedURI] = useState<string>('')

  const navigate = useNavigate()

  useEffect(() => {
    const storedUserProfile = localStorage.getItem('userProfile')
    if (!storedUserProfile) {
      return
    }

    const userProfile = JSON.parse(storedUserProfile) as User
    if (!userProfile) return

    setUserProfileData(userProfile)
    setEncodedURI(
      `/teammedlemmer/${encodeURIComponent(userProfile.principal_name ? userProfile.principal_name.split('@')[0] : userProfile.email.split('@')[0])}`
    )
    setFallbackInitials(userProfile.first_name[0] + userProfile.last_name[0])

    const base64Image = userProfile?.photo
    if (!base64Image) return
    setImageSrc(`data:image/png;base64,${base64Image}`)
  }, [])

  const handleClick = () => {
    if (encodedURI === '') return
    navigate(encodedURI)
  }

  return (
    <div className={styles.avatar} onClick={handleClick}>
      {imageSrc ? (
        <img src={imageSrc} alt='User' />
      ) : (
        <div className={styles.initials}>{userProfileData ? `${fallbackInitials}` : '??'}</div>
      )}
    </div>
  )
}
