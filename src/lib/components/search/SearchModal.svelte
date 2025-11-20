<script lang="ts">
	import { graphql } from '$houdini';
	import { envTagVariant } from '$lib/envTagVariant';
	import { Modal } from '@nais/ds-svelte-community';
	import {
		BriefcaseClockIcon,
		BucketIcon,
		DatabaseIcon,
		PackageIcon,
		PersonGroupIcon
	} from '@nais/ds-svelte-community/icons';
	import Search from './Search.svelte';

	let { open = $bindable() }: { open: boolean } = $props();

	const store = graphql(`
		query SearchQuery($query: String!, $type: SearchType) {
			search(first: 20, filter: { query: $query, type: $type }) {
				nodes {
					__typename
					... on Team {
						slug
						purpose
					}
				}
			}
		}
	`);

	const categories = {
		Team: {
			icon: PersonGroupIcon,
			urlName: 'team',
			prefix: 'team',
			type: 'TEAM'
		}
	} as const;

	let query = $state('');

	$effect(() => {
		if (query) {
			const timeout = setTimeout(() => {
				const [prefix, q] = query.split(':');
				const category = Object.values(categories).find((c) => c.prefix === prefix);
				const type = category?.type;
				store.fetch({ variables: { query: type ? q.trim() : query, type } });
			}, 300);

			return () => clearTimeout(timeout);
		}
	});
</script>

<Modal width="medium" bind:open class="search-modal">
	<Search
		close={() => (open = false)}
		bind:query
		loading={$store.fetching}
		results={query
			? $store.data?.search.nodes.map((result) => {
					const { icon, urlName } = categories[result.__typename];

					if (result.__typename === 'Team') {
						return {
							icon,
							label: result.slug,
							description: result.purpose,
							href: `/team/${result.slug}`,
							type: 'link'
						};
					}

					return {
						icon,
						label: result.name,
						description: result.team.slug,
						tag: {
							label: result.teamEnvironment.environment.name,
							variant: envTagVariant(result.teamEnvironment.environment.name)
						},
						href: `/team/${result.team.slug}/${result.teamEnvironment.environment.name}/${urlName}/${result.name}`,
						type: 'link'
					};
				})
			: undefined}
	/>
</Modal>
