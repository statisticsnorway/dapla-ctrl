type Paginated = {
	pageInfo: {
		totalCount: number;
	};
};

export const menuItems = ({
	path,
	member,
	isAdmin,
	inventory
}: {
	path: string;
	member: boolean;
	isAdmin: boolean;
	inventory?: {
		members: Paginated;
		sharedBuckets: Paginated;
	};
}): { label: string; href: string; active?: boolean; count?: number }[][] => {
	const split = path.split('/');

	const getInventory = (pageName?: string) => {
		const pageNameToInventory = {
			'shared-data': 'sharedBuckets',
			members: 'members'
		} as const;
		return {
			count:
				inventory?.[pageNameToInventory[pageName as keyof typeof pageNameToInventory]]?.pageInfo
					.totalCount
		};
	};

	const item =
		(baseUrl: string, page: string) =>
		(label: string, pageName?: string, matchSubPath?: string) => {
			const href = pageName ? `${baseUrl}/${pageName}` : baseUrl;
			const { count } = getInventory(pageName);
			const active =
				(matchSubPath && path.startsWith(`/team/${team}/${page}/${matchSubPath}/`)) ||
				pageName === page;

			return {
				label,
				href,
				...(active ? { active } : {}),
				...(count ? { count } : {})
			};
		};
	const [, , team, page] = split;
	const menuItem = item(`/team/${team}`, page);
	return [
		[menuItem('Oversikt')],
		[
			menuItem('Medlemmer', 'members'),
			menuItem('Datadeling', 'shared-data'),
			(member || isAdmin) && menuItem('Aktivitetslogg', 'activity-log')
		].filter(Boolean) as { label: string; href: string; active?: boolean }[]
	];
};
