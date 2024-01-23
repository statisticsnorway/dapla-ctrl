import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { User } from '../../api/UserApi';
import styles from './avatar.module.scss';

interface PageLayoutProps {
    fullName: string
}

export default function Avatar({ fullName }: PageLayoutProps) {
    const [userProfileData, setUserProfileData] = useState<User>();
    const [imageSrc, setImageSrc] = useState<string>();

    const fallbackInitials = `${fullName.split(' ')[0][0]}${fullName.split(' ')[1][0]}`

    const navigate = useNavigate();
    const encodedURI = encodeURI(`/teammedlemmer/${fullName}`);

    useEffect(() => {
        const storedUserProfile = localStorage.getItem('userProfile');
        if (!storedUserProfile) {
            return;
        }

        const userProfile = JSON.parse(storedUserProfile) as User;
        if (!userProfile) return;
        setUserProfileData(userProfile);

        const base64Image = userProfile?.photo;
        if (!base64Image) return;
        setImageSrc(`data:image/png;base64,${base64Image}`);
    }, []);

    const handleClick = () => {
        navigate(encodedURI);
    };

    return (
        <div className={styles.avatar} onClick={handleClick}>
            {imageSrc ? (
                <img src={imageSrc} alt="User" />
            ) : (
                <div className={styles.initials}>
                    {userProfileData ? `${fallbackInitials}` : '??'}
                </div>
            )}
        </div>
    );
}
