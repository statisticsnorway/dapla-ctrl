import { useEffect } from 'react';
import PageLayout from '../components/PageLayout/PageLayout'
import { useNavigate } from 'react-router-dom';

export default function Logout() {

    const navigate = useNavigate();

    useEffect(() => {
        localStorage.removeItem('token');
        navigate('/login');
    }, []);

    return (
        <PageLayout
            title="Du blir avlogget"
        />
    )
}