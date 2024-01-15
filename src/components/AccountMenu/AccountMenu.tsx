import styles from './accountmenu.module.scss'
import { useNavigate } from 'react-router-dom';

interface PageLayoutProps {
    fullName: string,
    image?: string,
}

export default function AccountMenu({ fullName, image }: PageLayoutProps) {
    // If the name does not contain a lastname, we only want to show the first letter of the firstname
    // else we want to show the first letter of both the firstname and the lastname
    const nameSplit = fullName.split(' ').filter((name) => name !== '');
    const initials = nameSplit.length <= 1 ? `${nameSplit[0][0]}` : `${nameSplit[0][0]}${nameSplit[1][0]}`;

    const navigate = useNavigate();

    const onClick = () => {
        const path_to_go = encodeURI(`/teammedlemmer/${fullName}`);
        navigate(path_to_go);
    }

    return (
        <div
            className={styles.AccountMenu}
            onClick={onClick}
        >
            {image ? <img src={image} /> : <div className={styles.initials}>{initials}</div>}
        </div>
    )
}