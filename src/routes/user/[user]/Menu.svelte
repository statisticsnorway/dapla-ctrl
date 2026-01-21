<script lang="ts">
	import { page } from '$app/state';
	import Menu from '$lib/ui/Menu.svelte';

	const menuItems = ({
		path
	}: {
		path: string;
	}): { label: string; href: string; active?: boolean; count?: number }[][] => {
		const split = path.split('/');

		const item =
			(baseUrl: string, page: string) =>
			(label: string, pageName?: string, matchSubPath?: string) => {
				const href = pageName ? `${baseUrl}/${pageName}` : baseUrl;
				const active =
					(matchSubPath && path.startsWith(`/user/${user}/${page}/${matchSubPath}/`)) ||
					pageName === page;

				return {
					label,
					href,
					...(active ? { active } : {})
				};
			};
		const [, , user, page] = split;
		const menuItem = item(`/user/${user}`, page);
		return [[menuItem('Brukeroversikt'), menuItem('Datadeling', 'shared-data')]];
	};
</script>

<Menu
	items={menuItems({
		path: page.url.pathname
	})}
/>
