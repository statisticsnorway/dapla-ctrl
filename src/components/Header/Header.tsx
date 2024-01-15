import styles from './header.module.scss'

import { Link } from '@statisticsnorway/ssb-component-library';
import AccountMenu from '../AccountMenu/AccountMenu';
import { useNavigate } from 'react-router-dom';

export default function Header() {

    let token = localStorage.getItem('token');
    if (token)
        var decoded_jwt = JSON.parse(atob(token.split('.')[1]));

    const navigate = useNavigate();
    return (
        <div className={styles.header}>
            <h2 style={{ "cursor": "pointer" }} onClick={() => navigate("/")}>Dapla ctrl</h2>
            <div className={styles.navigation}>
                <Link href="/teammedlemmer">Teammedlemmer</Link>
                {token ? <AccountMenu
                    firstName={decoded_jwt.given_name}
                    lastName={decoded_jwt.family_name}
                /> : <AccountMenu firstName='#' lastName='#' />} {/* What to use as fallback?*/}


            </div>

        </div>
    )
}