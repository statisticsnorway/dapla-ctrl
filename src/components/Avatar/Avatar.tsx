import styles from './avatar.module.scss';
import { useNavigate } from 'react-router-dom';

interface PageLayoutProps {
    fullName: string,
    image?: string,
}

export default function Avatar({ fullName }: PageLayoutProps) {
    const navigate = useNavigate();
    const handleClick = () => {
        const path_to_go = encodeURI(`/teammedlemmer/${fullName}`);
        navigate(path_to_go);
    };

    const userProfile = JSON.parse(localStorage.getItem('userProfile') || '{}');
    const base64Image = userProfile?.photo;
    const imageSrc = base64Image ? `data:image/png;base64,${base64Image}` : null;

    return (
        <div className={styles.Avatar} onClick={handleClick}>
            {imageSrc ? <img src={imageSrc} alt="User" /> :
                <div className={styles.initials}>{`${userProfile.firstName[0]}${userProfile.lastName[0]}`}</div>}
        </div>
    );
}
