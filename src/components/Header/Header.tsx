import styles from './header.module.scss'

import { Link } from '@statisticsnorway/ssb-component-library';
import Avatar from '../Avatar/Avatar';
import { useNavigate } from 'react-router-dom';
import { useState, useEffect } from 'react';

export default function Header(props: { isLoggedIn: boolean }) {
    const { isLoggedIn } = props
    const [accessToken, setAccessToken] = useState<string | null>(null);
    const [fullName, setUserName] = useState<string | null>(null);

    const navigate = useNavigate();
    useEffect(() => {
        const token = localStorage.getItem('access_token');
        if (!token) return;

        setAccessToken(token);

        const jwt = JSON.parse(atob(token.split('.')[1]));

        if (!jwt) return;
        setUserName(jwt.name);
    }, [isLoggedIn]);

    return (
        <div className={styles.header}>
            <h2 className={styles.title} onClick={() => navigate("/")}>Dapla ctrl</h2>
            {isLoggedIn &&
                <div className={styles.navigation}>
                    <Link href="/teammedlemmer">Teammedlemmer</Link>
                    {accessToken && <Avatar
                        fullName={fullName ? fullName : '??'}
                    />}
                </div>
            }
        </div>
    )
}