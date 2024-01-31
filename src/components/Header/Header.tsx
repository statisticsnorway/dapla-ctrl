import styles from './header.module.scss'

import { Link } from '@statisticsnorway/ssb-component-library'
import Avatar from '../Avatar/Avatar'
import { useNavigate } from 'react-router-dom'

export default function Header() {
  const navigate = useNavigate()

  return (
    <div className={styles.header}>
      <h2 className={styles.title} onClick={() => navigate('/')}>
        Dapla ctrl
      </h2>
      <div className={styles.navigation}>
        <Link href='/teammedlemmer'>Teammedlemmer</Link>
        <Avatar />
      </div>
    </div>
  )
}
