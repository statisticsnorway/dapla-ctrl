<script lang="ts">
	import { page } from '$app/state';
	import { graphql } from '$houdini';
	import Menu from '$lib/ui/Menu.svelte';
	import { setInventoryRefetcher } from './userContext.svelte';

	const {
		user
	}: {
		user: string;
	} = $props();

	type Paginated = {
		pageInfo: { totalCount: number };
	};

	const menuItems = ({
		path,
		inventory
	}: {
		path: string;
		inventory?: {
			sharedBucketsAccess: Paginated;
		};
	}): { label: string; href: string; active?: boolean; count?: number }[][] => {
		const split = path.split('/');

		const getInventory = (pageName?: string) => {
			const pageNameToINventory = {
				'shared-data': 'sharedBucketsAccess'
			} as const;
			return {
				count:
					inventory?.[pageNameToINventory[pageName as keyof typeof pageNameToINventory]]?.pageInfo
						.totalCount
			};
		};

		const item =
			(baseUrl: string, page: string) =>
			(label: string, pageName?: string, matchSubPath?: string) => {
				const href = pageName ? `${baseUrl}/${pageName}` : baseUrl;
				const { count } = getInventory(pageName);
				const active =
					(matchSubPath && path.startsWith(`/user/${user}/${page}/${matchSubPath}/`)) ||
					pageName === page;

				return {
					label,
					href,
					...(active ? { active } : {}),
					...(count ? { count } : {})
				};
			};
		const [, , user, page] = split;
		const menuItem = item(`/user/${user}`, page);
		return [[menuItem('Brukeroversikt'), menuItem('Datadeling', 'shared-data')]];
	};

	const Inventory = $derived(
		graphql(`
			query UserInventory($user: String) @cache(policy: CacheAndNetwork) {
				user(email: $user) {
					sharedBucketsAccess(first: 1) {
						pageInfo {
							totalCount
						}
					}
				}
			}
		`)
	);

	$effect(() => {
		Inventory.fetch({
			variables: {
				user: user
			}
		});
	});

	setInventoryRefetcher(() => {
		Inventory.fetch({
			variables: {
				user: user
			}
		});
	});
</script>

<Menu
	items={menuItems({
		path: page.url.pathname,
		inventory: $Inventory.fetching ? undefined : $Inventory.data?.user
	})}
/>
