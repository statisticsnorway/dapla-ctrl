<script lang="ts">
	import type { PageProps } from './$types';
	import TeamMembersTable from './TeamMembersTable.svelte';
	import { changeParams } from '$lib/utils/searchparams';
	import type { MyTeamMembers$result } from '$houdini';
	import type { TeamMemberData } from './teamMembersUtils';

	let { data }: PageProps = $props();
	let { MyTeamMembers } = $derived(data);

	type TeamMember = Extract<
		MyTeamMembers$result['me'],
		{ __typename: 'User' }
	>['teamMembers']['nodes'][0];

	const changeQuery = (params: { after?: string; before?: string }) => {
		changeParams(params);
	};

	function transformTeamMembersData(u: TeamMember): TeamMemberData {
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
	}
</script>

{#if $MyTeamMembers.fetching}
	<p>Laster...</p>
{:else if $MyTeamMembers.data?.me.__typename === 'User'}
	<TeamMembersTable
		teamMembers={$MyTeamMembers.data.me.teamMembers.nodes.map(transformTeamMembersData)}
		pageInfo={$MyTeamMembers.data.me.teamMembers.pageInfo}
		loaders={{
			loadPreviousPage: () => {
				changeQuery({
					after: '',
					before:
						$MyTeamMembers.data?.me.__typename === 'User'
							? ($MyTeamMembers.data?.me.teamMembers.pageInfo.startCursor ?? '')
							: ''
				});
			},
			loadNextPage: () => {
				changeQuery({
					before: '',
					after:
						$MyTeamMembers.data?.me.__typename === 'User'
							? ($MyTeamMembers.data?.me.teamMembers.pageInfo.endCursor ?? '')
							: ''
				});
			}
		}}
	/>
{/if}
