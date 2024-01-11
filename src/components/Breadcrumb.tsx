import { Breadcrumb as OriginalBreadcrumb } from '@statisticsnorway/ssb-component-library';
import { useLocation } from 'react-router-dom';

function Breadcrumb() {
    const location = useLocation();
    const pathnames = location.pathname.split('/').filter(x => x);

    const breadcrumbItems = pathnames.map((value, index) => {
        const last = index === pathnames.length - 1;
        const to = `/${pathnames.slice(0, index + 1).join('/')}`;

        return {
            text: value.charAt(0).toUpperCase() + value.slice(1),
            link: last ? undefined : to
        };
    });

    // Determine the items to pass to OriginalBreadcrumb
    const items = location.pathname === '/'
        ? [{ text: 'Forsiden' }]
        : [{ text: 'Forsiden', link: '/' }, ...breadcrumbItems];

    return (
        <div className="container">
            <OriginalBreadcrumb items={items} />
        </div>
    );
}

export default Breadcrumb;