import { useContext, useEffect, useState } from 'react';
import { Breadcrumb as OriginalBreadcrumb } from '@statisticsnorway/ssb-component-library';
import { useLocation } from 'react-router-dom';
import { DaplaCtrlContext } from '../provider/DaplaCtrlProvider';

export default function Breadcrumb() {
    const [displayName, setDisplayName] = useState<string>();
    const location = useLocation();
    const pathnames = location.pathname.split('/').filter(x => x).map(x => decodeURI(x));

    const { breadcrumbUserProfileDisplayName } = useContext(DaplaCtrlContext);

    useEffect(() => {
        if (pathnames[0] === 'teammedlemmer' && pathnames.length > 1) {
            if (breadcrumbUserProfileDisplayName) {
                setDisplayName(breadcrumbUserProfileDisplayName);
            }
        }
    }, [location, breadcrumbUserProfileDisplayName]);


    const breadcrumbItems = pathnames.map((value, index) => {
        const last = index === pathnames.length - 1;
        const to = `/${pathnames.slice(0, index + 1).join('/')}`;

        let displayValue = value.charAt(0).toUpperCase() + value.slice(1);
        if (index === 1 && displayName && pathnames[0] === 'teammedlemmer') {
            displayValue = displayName;
        }

        return {
            text: displayValue,
            link: last ? undefined : to
        };
    });

    const items = location.pathname === '/'
        ? [{ text: 'Forsiden' }]
        : [{ text: 'Forsiden', link: '/' }, ...breadcrumbItems];

    return (
        <OriginalBreadcrumb items={items} />
    );
}
