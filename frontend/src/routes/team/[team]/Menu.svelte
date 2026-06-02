<script lang="ts">
	import { page } from '$app/state';
	import { menuItems } from '$lib/menuItems';
	import Menu from '$lib/ui/Menu.svelte';
	import { graphql } from '$houdini';
	import { setInventoryRefetcher } from './teamContext.svelte';

	const {
		member,
		isAdmin,
		teamSlug
	}: {
		member: boolean;
		isAdmin: boolean;
		teamSlug: string;
	} = $props();

	const Inventory = $derived(
		graphql(`
			query Inventory($team: Slug!) @cache(policy: CacheAndNetwork) {
				team(slug: $team) {
					members(first: 1) {
						pageInfo {
							totalCount
						}
					}
					sharedBuckets(first: 1) {
						pageInfo {
							totalCount
						}
					}
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
				team: teamSlug
			}
		});
	});

	setInventoryRefetcher(() => {
		Inventory.fetch({
			variables: {
				team: teamSlug
			}
		});
	});
</script>

<Menu
	items={menuItems({
		path: page.url.pathname,
		member,
		isAdmin,
		inventory: $Inventory.fetching ? undefined : $Inventory.data?.team
	})}
/>
