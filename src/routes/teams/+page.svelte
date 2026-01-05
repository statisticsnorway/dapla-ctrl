<script lang="ts">
	import type { PageProps } from './$types';
	import { type TeamsData } from '../TeamsTable.svelte';
	import type { AllTeams$result } from '$houdini';
	import TeamOverview from '$lib/domain/TeamOverview.svelte';

	let { data }: PageProps = $props();

	let { AllTeams, UserInfo } = $derived(data);

	const canCreateTeam = $derived.by(() => {
		let me = $UserInfo.data?.me;
		if (me?.__typename !== 'User') return false;
		return me?.isAdmin || me?.isSectionManager;
	});

	let userTeamsCount = $derived(
		($AllTeams.data?.me.__typename == 'User' && $AllTeams.data.me.teams?.pageInfo.totalCount) || 0
	);

	type TeamNode = AllTeams$result['teams']['nodes'][0];

	let allTeamsCount = $derived($AllTeams.data?.teams?.nodes.length || 0);
	function transformTeamData(teamNode: TeamNode): TeamsData {
		const team = teamNode;
		const manager = team.section.manager;

		return {
			slug: team.slug,
			displayName: team.displayName,
			purpose: team.purpose,
			memberCount: team.members.pageInfo.totalCount,
			manager: {
				name: manager?.name ?? 'Mangler seksjonsleder',
				email: manager?.email ?? ''
			},
			section: {
				code: team.section.code,
				name: team.section.name
			}
		};
	}
</script>

<svelte:head><title>Alle team - Dapla Ctrl</title></svelte:head>

<div class="page">
	<TeamOverview
		{canCreateTeam}
		{userTeamsCount}
		{allTeamsCount}
		teamsData={$AllTeams.data?.teams.nodes.map(transformTeamData) ?? []}
		teamTableDefaultFields={data.teamTableFields}
		pageInfo={$AllTeams.data?.teams.pageInfo ?? undefined}
		loaders={{
			loadPreviousPage: () => AllTeams.loadPreviousPage(),
			loadNextPage: () => AllTeams.loadNextPage()
		}}
	/>
</div>

<style>
	.page {
		margin-inline: var(--margin-default);
	}
</style>
