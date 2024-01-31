import styles from './header.module.scss'

import { Link } from '@statisticsnorway/ssb-component-library'
import Avatar from '../Avatar/Avatar'
import { useNavigate } from 'react-router-dom'

export default function Header(props: { isLoggedIn: boolean }) {
  const { isLoggedIn } = props
  const navigate = useNavigate()

  return (
    <div className={styles.header}>
      <h2 className={styles.title} onClick={() => navigate('/')}>
        Dapla ctrl
      </h2>
      {isLoggedIn && (
        <div className={styles.navigation}>
          <Link href='/teammedlemmer'>Teammedlemmer</Link>
          <Avatar />
        </div>
      )}
    </div>
  )
}
