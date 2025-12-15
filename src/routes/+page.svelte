<script lang="ts">
	import Pagination from '$lib/ui/Pagination.svelte';
	import { Button, Heading } from '@nais/ds-svelte-community';
	import type { PageProps } from './$types';
	import TeamsTable, { type TeamsData } from './TeamsTable.svelte';
	import type { UserTeams$result } from '$houdini';

	let { data }: PageProps = $props();

	let { UserTeams, UserInfo } = $derived(data);

	const canCreateTeam = $derived.by(() => {
		let me = $UserInfo.data?.me;
		if (me?.__typename !== 'User') return false;
		return me?.isAdmin || me?.isSectionManager;
	});

	type TeamNode = Extract<UserTeams$result['me'], { __typename: 'User' }>['teams']['nodes'][0];
	type GroupNode = TeamNode['team']['groups']['nodes'][0];
	type GroupUserNode = GroupNode['members']['nodes'][0];

	let userTeamsCount = $derived(
		$UserTeams.data?.me.__typename == 'User' && $UserTeams.data.me.teams?.nodes.length
	);

	function transformTeamData(teamNode: TeamNode): TeamsData {
		const team = teamNode;

		const allManagers = team.team.groups.nodes
			.filter((group: GroupNode) => group.category === 'managers' && !group.suffix)
			.flatMap((managerGroup: GroupNode) =>
				managerGroup.members.nodes.map((member: GroupUserNode) => member.user)
			);
		const uniqueManagers = Array.from(
			new Map(allManagers.map((user: GroupUserNode['user']) => [user.email, user])).values()
		);

		return {
			slug: team.team.slug,
			memberCount: team.team.members.pageInfo.totalCount,
			managers: uniqueManagers,
			section: {
				code: team.team.section.code,
				name: team.team.section.name
			}
		};
	}
</script>

<svelte:head><title>Dapla Ctrl</title></svelte:head>

<div class="page">
	<div class="content-wrapper">
		<div class="header">
			<Heading level="1" size="large">Teamoversikt</Heading>
			{#if canCreateTeam}
				<Button as="a" size="medium" href="/team/create" variant="primary">Opprett team</Button>
			{/if}
		</div>
		{#if $UserTeams.data}
			{#if $UserTeams.data.me.__typename == 'User'}
				<div class="container">
					<div>
						<div class="section-header">
							<Heading level="2" spacing>Mine team ({userTeamsCount})</Heading>
							<Button as="a" size="small" href="/teams" variant="secondary">Se alle team</Button>
						</div>

						<TeamsTable teamsData={$UserTeams.data.me.teams.nodes.map(transformTeamData)} />
					</div>
				</div>
				<Pagination
					page={$UserTeams.data.me.teams.pageInfo}
					loaders={{
						loadPreviousPage: () => UserTeams.loadPreviousPage(),
						loadNextPage: () => UserTeams.loadNextPage()
					}}
				/>
			{/if}
		{/if}
	</div>
</div>

<style>
	.page {
		padding-top: 4rem;
	}
	.content-wrapper {
		background: var(--ax-bg-default);
		position: relative;
		top: -40px;
		padding: var(--ax-space-24);
		border-radius: 12px;
		max-width: 900px;
		margin-inline: auto;
	}

	.header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--ax-space-16);
	}

	.container {
		margin-top: var(--spacing-layout);
		display: flex;
		flex-direction: column;
		gap: var(--spacing-layout);
	}

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--ax-space-16);
	}
</style>
