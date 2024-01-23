import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { User } from '../../api/UserApi';
import styles from './avatar.module.scss';

interface PageLayoutProps {
    fullName: string,
    image?: string,
}

export default function Avatar({ fullName }: PageLayoutProps) {
    const [userProfileData, setUserProfileData] = useState<User>();
    const [imageSrc, setImageSrc] = useState<string>("");

    const navigate = useNavigate();
    const encodedURI = encodeURI(`/teammedlemmer/${fullName}`);

    useEffect(() => {
        const storedUserProfile = localStorage.getItem('userProfile');
        if (!storedUserProfile) {
            navigate('/login', { state: { from: encodedURI } });
            return;
        }

        const userProfile = JSON.parse(storedUserProfile) as User;
        const base64Image = userProfile?.photo;
        setUserProfileData(userProfile);
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
                    {userProfileData ? `${userProfileData.firstName?.[0]}${userProfileData.lastName?.[0]}` : '??'}
                </div>
            )}
        </div>
    );
}
