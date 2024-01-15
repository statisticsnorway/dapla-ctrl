import styles from './accountmenu.module.scss'
import { useNavigate } from 'react-router-dom';
import { useState, useEffect, useRef } from 'react';

interface PageLayoutProps {
    firstName: string,
    lastName: string,
    image?: string,
}

export default function AccountMenu({ firstName, lastName, image }: PageLayoutProps) {
    const initials = `${firstName[0]}${lastName[0]}`;
    const navigate = useNavigate();
    const [visibleMenu, setVisibleMenu] = useState(false);
    const menuRef = useRef<HTMLDivElement>(null);

    // If the user clicks outside of the menu, close it
    useEffect(() => {
        function handleClickOutside(event: MouseEvent) {
            if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
                setVisibleMenu(false);
            }
        }

        document.addEventListener("mousedown", handleClickOutside);
        return () => {
            document.removeEventListener("mousedown", handleClickOutside);
        };
    }, [menuRef]);

    const onClick = () => {
        const path_to_go = encodeURI(`/teammedlemmer/${firstName} ${lastName}`);
        navigate(path_to_go);
        setVisibleMenu(!visibleMenu);
    }

    return (
        <div
            ref={menuRef}
            className={styles.AccountMenu}
            onClick={onClick}
        >
            {image ? <img src={image} alt="Gå til profil" /> : <div className={styles.initials}>{initials}</div>}
        </div>
    )
}