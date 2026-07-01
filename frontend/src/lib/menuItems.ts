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
		sharedBucketsAccess: Paginated;
	};
}): { label: string; href: string; active?: boolean; count?: number }[][] => {
	const split = path.split('/');

	const getInventory = (pageName?: string) => {
		let count: number | undefined = undefined;
		switch (pageName) {
			case 'members':
				count = inventory?.members.pageInfo.totalCount;
				break;
			case 'shared-data':
				count = inventory
					? inventory.sharedBuckets.pageInfo.totalCount +
						inventory.sharedBucketsAccess.pageInfo.totalCount
					: undefined;
				break;
		}
		return {
			count: count
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
			isAdmin && menuItem('Dapla Lab', 'launch-lab'),
			(member || isAdmin) && menuItem('Aktivitetslogg', 'activity-log'),
			isAdmin && menuItem('Innstillinger', 'settings')
		].filter(Boolean) as { label: string; href: string; active?: boolean }[]
	];
};
