<script lang="ts">
	import type { PageProps } from './$types';
	import TeamMembersTable from './TeamMembersTable.svelte';
	import { type TeamMemberData } from './teamMembersUtils';
	import { changeParams } from '$lib/utils/searchparams';

	let { data }: PageProps = $props();
	let { MyTeamMembers } = $derived(data);

	const changeQuery = (params: { after?: string; before?: string }) => {
		changeParams(params);
	};

	let meUser = $MyTeamMembers.data?.me.__typename === 'User' ? $MyTeamMembers.data?.me : undefined;

	let teamMembers: TeamMemberData[] =
		meUser?.teamMembers.nodes.map((u) => {
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

{#if $MyTeamMembers.fetching}
	<p>Laster...</p>
{:else if $MyTeamMembers.data?.me.__typename === 'User'}
	<TeamMembersTable
		{teamMembers}
		pageInfo={$MyTeamMembers.data.me.teamMembers.pageInfo}
		loaders={{
			loadPreviousPage: () => {
				changeQuery({
					before:
						$MyTeamMembers.data?.me.__typename === 'User'
							? ($MyTeamMembers.data?.me.teamMembers.pageInfo.startCursor ?? '')
							: ''
				});
			},
			loadNextPage: () => {
				changeQuery({
					after:
						$MyTeamMembers.data?.me.__typename === 'User'
							? ($MyTeamMembers.data?.me.teamMembers.pageInfo.endCursor ?? '')
							: ''
				});
			}
		}}
	/>
{/if}
