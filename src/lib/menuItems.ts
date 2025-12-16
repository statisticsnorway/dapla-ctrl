export const menuItems = ({
	path,
	member,
	isAdmin
}: {
	path: string;
	member: boolean;
	isAdmin: boolean;
}): { label: string; href: string; active?: boolean; count?: number }[][] => {
	const split = path.split('/');

	const item =
		(baseUrl: string, page: string) =>
		(label: string, pageName?: string, matchSubPath?: string) => {
			const href = pageName ? `${baseUrl}/${pageName}` : baseUrl;
			const active =
				(matchSubPath && path.startsWith(`/team/${team}/${page}/${matchSubPath}/`)) ||
				pageName === page;

			return {
				label,
				href,
				...(active ? { active } : {})
			};
		};
	const [, , team, page] = split;
	const menuItem = item(`/team/${team}`, page);
	return [
		[menuItem('Teamoversikt')],
		[
			menuItem('Grupper', 'groups'),
			(member || isAdmin) && menuItem('Instillinger', 'settings'),
			(member || isAdmin) && menuItem('Aktivitetslogg', 'activity-log')
		].filter(Boolean) as { label: string; href: string; active?: boolean }[]
	];
};
