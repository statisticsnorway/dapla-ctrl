<script lang="ts">
	import { browser } from '$app/environment';
	import { page } from '$app/state';
	import { graphql } from '$houdini';
	import { get } from 'svelte/store';
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
					(matchSubPath && path.startsWith(`/member/${user}/${page}/${matchSubPath}/`)) ||
					pageName === page;

				return {
					label,
					href,
					...(active ? { active } : {}),
					...(count ? { count } : {})
				};
			};
		const [, , user, page] = split;
		const menuItem = item(`/member/${user}`, page);
		return [
			[
				menuItem('Oversikt'),
				menuItem('Medlemskap', 'membership'),
				menuItem('Datatilgang', 'shared-data')
			]
		];
	};

	const Inventory = $derived.by(() => {
		if (!browser) return null;
		return graphql(`
			query UserInventory($user: String) @cache(policy: CacheAndNetwork) {
				user(email: $user) {
					sharedBucketsAccess(first: 1) {
						pageInfo {
							totalCount
						}
					}
				}
			}
		`);
	});

	$effect(() => {
		if (Inventory) {
			Inventory.fetch({
				variables: {
					user: user
				}
			});
		}
	});

	setInventoryRefetcher(() => {
		if (Inventory) {
			Inventory.fetch({
				variables: {
					user: user
				}
			});
		}
	});

	const inventoryData = $derived.by(() => {
		if (!Inventory) return undefined;
		const store = Inventory as NonNullable<typeof Inventory>;
		const storeValue = get(store);
		return !storeValue.fetching ? storeValue.data?.user : undefined;
	});
</script>

<Menu
	items={menuItems({
		path: page.url.pathname,
		inventory: inventoryData
	})}
/>
