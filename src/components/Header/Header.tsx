import styles from './header.module.scss'

import { Link } from '@statisticsnorway/ssb-component-library';
import Avatar from '../Avatar/Avatar';
import { useNavigate } from 'react-router-dom';
import {
    jwtRegex
} from '../../utils/regex';

export default function Header(props: { isLoggedIn: boolean }) {
    const { isLoggedIn } = props

    const token = localStorage.getItem('access_token');
    let decoded_jwt;

    if (token && jwtRegex.test(token))
        decoded_jwt = JSON.parse(atob(token.split('.')[1]));

    const navigate = useNavigate();
    return (
        <div className={styles.header}>
            <h2 className={styles.title} onClick={() => navigate("/")}>Dapla ctrl</h2>
            {isLoggedIn &&
                <div className={styles.navigation}>
                    <Link href="/teammedlemmer">Teammedlemmer</Link>
                    {token && <Avatar
                        fullName={decoded_jwt.name}
                    />}
                </div>
            }
        </div>
    )
}