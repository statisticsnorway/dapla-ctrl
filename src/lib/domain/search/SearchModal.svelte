<script lang="ts">
	import { graphql } from '$houdini';
	import { Modal } from '@nais/ds-svelte-community';
	import {
		PersonGroupIcon,
		HexagonGridIcon,
		FloppydiskIcon
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
					}
					... on Group {
						name
						teamSlug
						category
						members {
							pageInfo {
								totalCount
							}
						}
					}
					... on SharedBucket {
						name
						shortName
						team {
							slug
						}
					}
				}
			}
		}
	`);

	const categories = {
		Team: {
			icon: HexagonGridIcon,
			urlName: 'team',
			prefix: 'team',
			type: 'TEAM'
		},
		Group: {
			icon: PersonGroupIcon,
			urlName: 'group',
			prefix: 'gruppe',
			type: 'GROUP'
		},
		SharedBucket: {
			icon: FloppydiskIcon,
			urlName: 'sharedbucket',
			prefix: 'delt',
			type: 'SHAREDBUCKET'
		}
	} as const;

	let query = $state('');

	$effect(() => {
		if (query) {
			const timeout = setTimeout(() => {
				const [prefix, q] = query.split(':');
				const category = Object.values(categories).find((c) => c.prefix === prefix);
				const type = category?.type;
				let searchQuery;
				if (type) {
					if (q && q.trim()) {
						searchQuery = q.trim();
					} else {
						searchQuery = '';
					}
				} else {
					searchQuery = query;
				}
				store.fetch({ variables: { query: searchQuery, type } });
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
					const { icon } = categories[result.__typename];
					if (result.__typename === 'Team') {
						return {
							icon,
							label: result.slug,
							description: '',
							href: `/team/${result.slug}`,
							type: 'link'
						};
					} else if (result.__typename === 'Group') {
						const memberCount = result.members.pageInfo.totalCount;
						return {
							icon,
							label: result.name,
							description: `${memberCount} medlem${memberCount != 1 ? 'mer' : ''}`,
							href: `/team/${result.teamSlug}/groups`,
							type: 'link'
						};
					}
					// SharedBucket
					return {
						icon,
						label: result.shortName,
						description: result.name,
						href: `team/${result.team.slug}/shared-data/${result.name}`,
						type: 'link'
					};
				})
			: undefined}
	/>
</Modal>
