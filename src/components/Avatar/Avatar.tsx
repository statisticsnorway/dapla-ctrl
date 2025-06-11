import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { User } from '../../@types/user'
import styles from './avatar.module.scss'

const Avatar = () => {
  const [userProfileData, setUserProfileData] = useState<User>(null)
  const [imageSrc, setImageSrc] = useState<URL>(null)
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
    setEncodedURI(`/teammedlemmer/${userProfile.principal_name}`)
    setFallbackInitials(userProfile.first_name[0] + userProfile.last_name[0])

    const base64Image = userProfile?.photo
    if (!base64Image) return

    try {
      const byteCharacters = atob(base64Image)
      const byteNumbers = new Array(byteCharacters.length)
      for (let i = 0; i < byteCharacters.length; i++) {
        byteNumbers[i] = byteCharacters.charCodeAt(i)
      }

      const byteArray = new Uint8Array(byteNumbers)
      const blob: Blob = new Blob([byteArray], { type: 'image/png' })
      const blobUrl: URL = URL.createObjectURL(blob)

      setImageSrc(blobUrl)

      // Cleanup: revoke the blob URL when the component unmounts
      return () => {
        URL.revokeObjectURL(blobUrl)
      }
    } catch (error) {
      console.error('Failed to convert base64 string of the avatar photo to Blob URL', error)
    }
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

export default Avatar
