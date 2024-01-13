import styles from './header.module.scss'
import { Link } from 'react-router-dom';

export default function Header() {
    return (
        <div className={styles.header}>
            <span>Dapla ctrl</span>
            <Link to="/medlemmer">Medlemmer</Link>
        </div>
    )
}