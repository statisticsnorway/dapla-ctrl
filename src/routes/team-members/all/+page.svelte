<script lang="ts">
	import { changeParams } from '$lib/utils/searchparams';
	import type { PageProps } from './$types';
	import TeamMembersTable from '../TeamMembersTable.svelte';
	import type { TeamMemberData } from '../teamMembersUtils';
	import type { AllTeamMembers$result } from '$houdini';

	let { data }: PageProps = $props();
	let { AllTeamMembers } = $derived(data);

	type TeamMember = AllTeamMembers$result['users']['nodes'][0];

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

{#if $AllTeamMembers.data}
	<TeamMembersTable
		teamMembers={$AllTeamMembers.data?.users.nodes.map(transformTeamMembersData)}
		pageInfo={$AllTeamMembers.data.users.pageInfo}
		loaders={{
			loadPreviousPage: () => {
				changeQuery({ after: '', before: $AllTeamMembers.data?.users.pageInfo.startCursor ?? '' });
			},
			loadNextPage: () => {
				changeQuery({ before: '', after: $AllTeamMembers.data?.users.pageInfo.endCursor ?? '' });
			}
		}}
	/>
{:else}
	<p>Laster...</p>
{/if}
