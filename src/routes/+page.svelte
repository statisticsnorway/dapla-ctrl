<script lang="ts">
	import type { PageProps } from './$types';
	import { type TeamsData } from '$lib/domain/TeamOverview.svelte';
	import type { UserTeams$result } from '$houdini';
	import TeamOverview from '$lib/domain/TeamOverview.svelte';

	let { data }: PageProps = $props();

	let { UserTeams, UserInfo } = $derived(data);

	let userEmail = $derived(
		$UserInfo.data?.me.__typename === 'User' ? $UserInfo.data.me.email : undefined
	);

	const canCreateTeam = $derived.by(() => {
		let me = $UserInfo.data?.me;
		if (me?.__typename !== 'User') return false;
		// TODO: uncomment when we support creating teams
		return me?.isAdmin; //|| me?.isSectionManager;
	});

	type TeamNode = Extract<UserTeams$result['me'], { __typename: 'User' }>['teams']['nodes'][0];

	let userTeamsCount = $derived(
		($UserTeams.data?.me.__typename == 'User' && $UserTeams.data.me.teams?.nodes.length) || 0
	);

	let allTeamsCount = $derived($UserTeams.data?.teams?.pageInfo.totalCount || 0);

	function transformTeamData(teamNode: TeamNode): TeamsData {
		const team = teamNode;
		const manager = team.team.section.manager;
		return {
			id: team.team.id,
			slug: team.team.slug,
			displayName: team.team.displayName,
			memberCount: team.team.members.pageInfo.totalCount,
			isManaged: team.team.isManaged,
			manager: {
				name: manager?.name ?? 'Mangler seksjonsleder',
				email: manager?.email ?? ''
			},
			section: {
				code: team.team.section.code,
				name: team.team.section.name
			},
			userGroups: team.groups.filter((g) => g !== null).map((g) => g.name) ?? []
		};
	}
	let teamsData = $derived.by(() => {
		if ($UserTeams.data && $UserTeams.data.me.__typename == 'User') {
			return $UserTeams.data?.me.teams.nodes.map(transformTeamData) ?? [];
		}
		return [];
	});

	let pageInfo = $derived.by(() => {
		if ($UserTeams.data && $UserTeams.data.me.__typename == 'User') {
			return $UserTeams.data?.me.teams?.pageInfo ?? undefined;
		}
		return undefined;
	});
</script>

<svelte:head><title>Dapla Ctrl</title></svelte:head>

<div class="page">
	<TeamOverview
		{userEmail}
		{canCreateTeam}
		{userTeamsCount}
		{allTeamsCount}
		{teamsData}
		teamTableDefaultFields={data.teamTableFields}
		{pageInfo}
		loaders={{
			loadPreviousPage: () => UserTeams.loadPreviousPage(),
			loadNextPage: () => UserTeams.loadNextPage()
		}}
	/>
</div>

<style>
	.page {
		margin-inline: var(--margin-default);
	}
</style>
