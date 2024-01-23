import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import styles from './avatar.module.scss';

interface PageLayoutProps {
    fullName: string,
    image?: string,
    photo?: string,
}

export default function Avatar({ fullName, photo }: PageLayoutProps) {
    const [imageSrc, setImageSrc] = useState<string>();

    const fallbackInitials = `${fullName.split(' ')[0][0]}${fullName.split(' ')[1][0]}`

    const navigate = useNavigate();
    const encodedURI = encodeURI(`/teammedlemmer/${fullName}`);

    useEffect(() => {
        if (!photo) return;
        setImageSrc(`data:image/png;base64,${photo}`);
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
                    {fullName ? `${fallbackInitials}` : '??'}
                </div>
            )}
        </div>
    );
}
