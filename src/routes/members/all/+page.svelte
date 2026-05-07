<script lang="ts">
	import { changeParams } from '$lib/utils/searchparams';
	import type { PageProps } from './$types';
	import TeamMembersTable, { type TeamMemberData } from '../TeamMembersTable.svelte';
	import type { AllTeamMembers$result } from '$houdini';

	let { data }: PageProps = $props();
	let { AllTeamMembers } = $derived(data);

	type TeamMember = AllTeamMembers$result['teamMembers']['nodes'][0];

	const changeQuery = (params: { after?: string; before?: string }) => {
		changeParams(params);
	};

	function transformTeamMembersData(u: TeamMember): TeamMemberData {
		return {
			id: u.id,
			user: {
				name: u.name,
				email: u.email
			},
			teamCount: u.teams.pageInfo.totalCount,
			dataAdminCount: u.dataAdmins.pageInfo.totalCount,
			section: u.section ?? undefined
		};
	}
</script>

{#if $AllTeamMembers.data}
	<TeamMembersTable
		teamMembers={$AllTeamMembers.data?.teamMembers.nodes.map(transformTeamMembersData)}
		pageInfo={$AllTeamMembers.data.teamMembers.pageInfo}
		selected={data.teamMembersTableField}
		loaders={{
			loadPreviousPage: () => {
				changeQuery({
					after: '',
					before: $AllTeamMembers.data?.teamMembers.pageInfo.startCursor ?? ''
				});
			},
			loadNextPage: () => {
				changeQuery({
					before: '',
					after: $AllTeamMembers.data?.teamMembers.pageInfo.endCursor ?? ''
				});
			}
		}}
	/>
{:else}
	<p>Laster...</p>
{/if}
