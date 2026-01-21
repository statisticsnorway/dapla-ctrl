<script lang="ts">
	import { changeParams } from '$lib/utils/searchparams';
	import type { PageProps } from './$types';
	import TeamMembersTable from '../TeamMembersTable.svelte';
	import { calculateDataAdminCount, type TeamMemberData } from '../teamMembersUtils';

	let { data }: PageProps = $props();
	let { AllTeamMembers } = $derived(data);

	const changeQuery = (params: { after?: string; before?: string }) => {
		changeParams(params);
	};

	let teamMembers = $derived.by(() => {
		if (!$AllTeamMembers.data) {
			return [];
		}

		const members: TeamMemberData[] = [];

		for (const edge of $AllTeamMembers.data.users.edges) {
			const user = edge.node;

			if (user.teams.pageInfo.totalCount === 0) {
				continue;
			}

			const dataAdminCount = calculateDataAdminCount(user.teams.nodes);

			const sectionManager = user.section?.manager
				? {
						name: user.section.manager.name,
						email: user.section.manager.email
					}
				: null;

			members.push({
				user: {
					id: user.id,
					name: user.name,
					email: user.email
				},
				teamCount: user.teams.pageInfo.totalCount,
				dataAdminCount,
				sectionManager
			});
		}

		return members;
	});
</script>

{#if $AllTeamMembers.data}
	<TeamMembersTable
		{teamMembers}
		pageInfo={$AllTeamMembers.data.users.pageInfo}
		loaders={{
			loadPreviousPage: () => {
				changeQuery({ before: $AllTeamMembers.data?.users.pageInfo.startCursor ?? '' });
			},
			loadNextPage: () => {
				changeQuery({ after: $AllTeamMembers.data?.users.pageInfo.endCursor ?? '' });
			}
		}}
	/>
{:else}
	<p>Laster...</p>
{/if}
