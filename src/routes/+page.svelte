<script lang="ts">
	import Pagination from '$lib/ui/Pagination.svelte';
	import { Button, Heading } from '@nais/ds-svelte-community';
	import type { PageProps } from './$types';
	import TeamsTable, { type TeamsData } from './TeamsTable.svelte';
	import type { UserTeams$result } from '$houdini';
	import Tab from '$lib/ui/Tab.svelte';
	import Tabs from '$lib/ui/Tabs.svelte';

	let { data }: PageProps = $props();

	let { UserTeams, UserInfo } = $derived(data);

	const canCreateTeam = $derived.by(() => {
		let me = $UserInfo.data?.me;
		if (me?.__typename !== 'User') return false;
		return me?.isAdmin || me?.isSectionManager;
	});

	type TeamNode = Extract<UserTeams$result['me'], { __typename: 'User' }>['teams']['nodes'][0];

	let userTeamsCount = $derived(
		$UserTeams.data?.me.__typename == 'User' && $UserTeams.data.me.teams?.nodes.length
	);

	let allTeamsCount = $derived($UserTeams.data?.teams?.pageInfo.totalCount || 0);

	function transformTeamData(teamNode: TeamNode): TeamsData {
		const team = teamNode;
		const manager = team.team.section.manager;
		return {
			slug: team.team.slug,
			displayName: team.team.displayName,
			purpose: team.team.purpose,
			memberCount: team.team.members.pageInfo.totalCount,
			manager: {
				name: manager?.name ?? 'Mangler seksjonsleder',
				email: manager?.email ?? ''
			},
			section: {
				code: team.team.section.code,
				name: team.team.section.name
			},
			userGroups: team.groups?.filter((g) => g !== null).map((g) => g.name) ?? []
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
						<Tabs>
							<Tab href="/" active={true} title="Mine team ({userTeamsCount})" />
							<Tab href="/teams" active={false} title="Alle teams ({allTeamsCount})" />
						</Tabs>

						<TeamsTable
							teamsData={$UserTeams.data.me.teams.nodes.map(transformTeamData)}
							defaultSelected={data.teamTableFields}
						/>
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
</style>
