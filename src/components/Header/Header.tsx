import styles from './header.module.scss'

import { Link } from '@statisticsnorway/ssb-component-library';
import AccountMenu from '../AccountMenu/AccountMenu';
import { useNavigate } from 'react-router-dom';
import {
    jwtRegex
} from '../../utils/regex';

export default function Header(props: { isLoggedIn: boolean }) {
    const { isLoggedIn } = props

    let token = localStorage.getItem('token');
    if (token && jwtRegex.test(token))
        var decoded_jwt = JSON.parse(atob(token.split('.')[1]));

    const navigate = useNavigate();
    return (
        <div className={styles.header}>
            <h2 className={styles.title} onClick={() => navigate("/")}>Dapla ctrl</h2>
            {isLoggedIn &&
                <div className={styles.navigation}>
                    <Link href="/teammedlemmer">Teammedlemmer</Link>
                    {token ? <AccountMenu
                        fullName={decoded_jwt.name}
                    /> : <AccountMenu fullName='#' />}
                </div>
            }
        </div>
    )
}