import styles from './header.module.scss'

import { Link } from '@statisticsnorway/ssb-component-library'
import Avatar from '../Avatar/Avatar'
import { useNavigate } from 'react-router-dom'

import { BookOpen } from 'react-feather'

const Header = () => {
  const navigate = useNavigate()

  return (
    <div className={styles.header}>
      <h2 className={styles.title} onClick={() => navigate('/')}>
        Dapla ctrl
      </h2>
      <div className={styles.navigation}>
        <div className={styles.links}>
          <Link
            href='https://statistics-norway.atlassian.net/wiki/spaces/DAPLA/pages/3803611153/Dapla+Ctrl'
            isExternal={true}
            icon={<BookOpen size='20' />}
          >
            Dokumentasjon
          </Link>
          <Link href='/teammedlemmer'>Teammedlemmer</Link>
        </div>
        <Avatar />
      </div>
    </div>
  )
}

export default Header
