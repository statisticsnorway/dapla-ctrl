import styles from './header.module.scss'

import { Link } from 'react-router-dom';

export default function Header(props) {
    const { isLoggedIn } = props

    return (
        <div className={styles.header}>
            <span>Dapla ctrl</span>
            {isLoggedIn && <Link to="/medlemmer">Medlemmer</Link>}
        </div>
    )
}