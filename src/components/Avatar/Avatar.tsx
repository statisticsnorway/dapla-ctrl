import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { User } from '../../@types/user'
import styles from './avatar.module.scss'
import { Effect, Option as O } from 'effect'
import { useUserProfileStore } from '../../services/store'

// Convert base64 string to Blob URL while handling potential errors
const base64ToBlobUrl = (base64Image: string): Effect.Effect<string, Error> =>
  Effect.try({
    try: () => {
      const byteArray = new Uint8Array(Array.from(atob(base64Image), (c) => c.charCodeAt(0)))
      const blob = new Blob([byteArray], { type: 'image/png' })
      return URL.createObjectURL(blob)
    },
    catch: (unknownError) => new Error(`Failed to convert base64 avatar photo to Blob URL: ${unknownError.message}`),
  })

const Avatar = () => {
  const [imageSrc, setImageSrc] = useState<string>()
  const [fallbackInitials, setFallbackInitials] = useState<string>('??')
  const [encodedURI, setEncodedURI] = useState<string>('')
  const maybeUser: O.Option<User> = useUserProfileStore((state) => state.loggedInUser)

  const navigate = useNavigate()

  useEffect(() => {
    Effect.gen(function* () {
      const user: User = yield* O.match(maybeUser, {
        onNone: () => Effect.fail(new Error('User not logged in!')),
        onSome: (user) => Effect.succeed(user),
      })
      yield* Effect.sync(() => {
        setEncodedURI(`/teammedlemmer/${user.principal_name}`)
        setFallbackInitials(user.first_name[0] + user.last_name[0])
      })
      const base64Image = yield* O.match(O.fromNullable(user?.photo), {
        onNone: () => Effect.fail(new Error("User object doesn't contain photo")),
        onSome: (photo) => Effect.succeed(photo),
      })
      return yield* base64ToBlobUrl(base64Image)
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
