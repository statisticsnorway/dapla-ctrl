export const menuItems = ({
	path,
	member,
	isAdmin
}: {
	path: string;
	features?: object;
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
		[menuItem('Team Overview')],
		[
			menuItem('Groups', 'groups'),
			(member || isAdmin) && menuItem('Settings', 'settings'),
			member && menuItem('Activity Log', 'activity-log')
		].filter(Boolean) as { label: string; href: string; active?: boolean }[]
	];
};
