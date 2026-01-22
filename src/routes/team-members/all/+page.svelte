<script lang="ts">
	import { changeParams } from '$lib/utils/searchparams';
	import type { PageProps } from './$types';
	import TeamMembersTable from '../TeamMembersTable.svelte';
	import type { TeamMemberData } from '../teamMembersUtils';

	let { data }: PageProps = $props();
	let { AllTeamMembers } = $derived(data);

	const changeQuery = (params: { after?: string; before?: string }) => {
		changeParams(params);
	};

	let teamMembers: TeamMemberData[] =
		$AllTeamMembers.data?.users.nodes.map((u) => {
			return {
				user: {
					name: u.name,
					email: u.email
				},
				teamCount: u.teams.pageInfo.totalCount,
				dataAdminCount: u.dataAdmins.pageInfo.totalCount,
				sectionManager: u.section?.manager
					? {
							name: u.section.manager.name,
							email: u.section.manager.email
						}
					: undefined
			};
		}) ?? [];
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
