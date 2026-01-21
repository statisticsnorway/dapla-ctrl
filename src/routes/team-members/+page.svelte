<script lang="ts">
	import type { PageProps } from './$types';
	import TeamMembersTable from './TeamMembersTable.svelte';
	import TeamMembersActivityLog from './TeamMembersActivityLog.svelte';
	import { calculateDataAdminCount, type TeamMemberData } from './teamMembersUtils';

	let { data }: PageProps = $props();
	let { MyTeamMembers } = $derived(data);

	let teamMembers = $derived.by(() => {
		if (
			$MyTeamMembers.fetching ||
			!$MyTeamMembers.data ||
			$MyTeamMembers.data.me.__typename !== 'User'
		) {
			return [];
		}

		const members = $MyTeamMembers.data.me.teamMembers.nodes;
		const memberData: TeamMemberData[] = [];

		for (const user of members) {
			const dataAdminCount = calculateDataAdminCount(user.teams.nodes);
			const sectionManager = user.section?.manager
				? {
						name: user.section.manager.name,
						email: user.section.manager.email
					}
				: null;

			memberData.push({
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

		return memberData;
	});

	let userTeamSlugs = $derived.by(() => {
		if (
			$MyTeamMembers.fetching ||
			!$MyTeamMembers.data ||
			$MyTeamMembers.data.me.__typename !== 'User'
		) {
			return [];
		}

		const teamSlugs: string[] = [];
		for (const user of $MyTeamMembers.data.me.teamMembers.nodes) {
			for (const userTeam of user.teams.nodes) {
				if (!teamSlugs.includes(userTeam.team.slug)) {
					teamSlugs.push(userTeam.team.slug);
				}
			}
		}
		return teamSlugs;
	});
</script>

{#if $MyTeamMembers.fetching}
	<p>Laster...</p>
{:else if $MyTeamMembers.data && $MyTeamMembers.data.me.__typename === 'User'}
	<div class="main-layout">
		<div class="left-section">
			<TeamMembersTable {teamMembers} />
		</div>
		<div class="right-section">
			<TeamMembersActivityLog {teamMembers} {userTeamSlugs} />
		</div>
	</div>
{/if}

<style>
	.main-layout {
		display: grid;
		grid-template-columns: 1fr max-content;
		gap: var(--spacing-layout);
		align-items: start;
	}

	.left-section {
		display: flex;
		flex-direction: column;
		gap: var(--ax-space-4);
	}

	.right-section {
		max-width: 250px;
		display: flex;
		flex-direction: column;
		gap: var(--ax-space-16);
	}
</style>
