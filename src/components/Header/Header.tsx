import styles from './header.module.scss'

import { Link } from '@statisticsnorway/ssb-component-library'
import Avatar from '../Avatar/Avatar'
import { useNavigate } from 'react-router-dom'

import { Book, Users } from 'react-feather'

const Header = () => {
  const navigate = useNavigate()

  return (
    <div className={styles.header}>
      <h2 className={styles.title} onClick={() => navigate('/')}>
        Dapla Ctrl
      </h2>
      <div className={styles.navigation}>
        <div className={styles.links}>
          <Link href={import.meta.env.DAPLA_CTRL_DOCUMENTATION_URL ?? ''} isExternal={true} icon={<Book size='20' />}>
            Dokumentasjon
          </Link>
          <Link href='/teammedlemmer' icon={<Users size='20' />}>
            Teammedlemmer
          </Link>
        </div>
        <Avatar />
      </div>
    </div>
  )
}

export default Header
