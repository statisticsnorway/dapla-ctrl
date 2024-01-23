import styles from './header.module.scss'

import { Link } from '@statisticsnorway/ssb-component-library';
import Avatar from '../Avatar/Avatar';
import { useNavigate } from 'react-router-dom';
import { User } from '../../api/UserApi';
import { useEffect, useState } from 'react';

export default function Header(props: { isLoggedIn: boolean }) {
    const { isLoggedIn } = props
    const [userProfile, setUserProfile] = useState<User | null>(null);
    const navigate = useNavigate();

    useEffect(() => {
        const storedUserProfile = localStorage.getItem('userProfile');
        if (!storedUserProfile) {
            navigate('/login');
            return;
        }
        setUserProfile(JSON.parse(storedUserProfile));
    }, [isLoggedIn]);

    return (
        <div className={styles.header}>
            <h2 className={styles.title} onClick={() => navigate("/")}>Dapla ctrl</h2>
            {isLoggedIn &&
                <div className={styles.navigation}>
                    <Link href="/teammedlemmer">Teammedlemmer</Link>
                    {userProfile && <Avatar
                        fullName={userProfile.displayName.split(', ').reverse().join(' ')}
                        photo={userProfile.photo}
                    />}
                </div>
            }
        </div>
    )
}