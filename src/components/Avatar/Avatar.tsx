import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { UserProfile } from '../../@types/user'
import { UserNotLoggedIn } from '../../@types/error'
import styles from './avatar.module.scss'
import { Effect, Option as O } from 'effect'
import { useUserProfileStore } from '../../services/store'

const Avatar = () => {
  const [imageSrc, setImageSrc] = useState<string>()
  const [fallbackInitials, setFallbackInitials] = useState<string>('??')
  const [encodedURI, setEncodedURI] = useState<string>('')
  const maybeLoggedInUser: O.Option<UserProfile> = useUserProfileStore((state) => state.loggedInUser)

  const navigate = useNavigate()

  useEffect(() => {
    Effect.gen(function* () {
      const user: UserProfile = yield* O.match(maybeLoggedInUser, {
        onNone: () => Effect.fail(new UserNotLoggedIn('Could not find UserProfile object in the zustand store!')),
        onSome: Effect.succeed,
      })
      yield* Effect.sync(() => {
        setEncodedURI(`/teammedlemmer/${user.principalName}`)
        setFallbackInitials(user.firstName[0] + user.lastName[0])
      })
      return user.photo
    })
      .pipe(Effect.runPromise)
      .then((blobUrl) => setImageSrc(blobUrl))
  }, [])

  const handleClick = () => {
    if (encodedURI === '') return
    navigate(encodedURI)
  }

  return (
    <div className={styles.avatar} onClick={handleClick}>
      {imageSrc ? <img src={imageSrc} alt='User' /> : <div className={styles.initials}>{fallbackInitials}</div>}
    </div>
  )
}

export default Avatar
