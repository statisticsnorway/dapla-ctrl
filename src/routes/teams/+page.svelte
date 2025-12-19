<script lang="ts">
	import Pagination from '$lib/ui/Pagination.svelte';
	import { Button, Heading } from '@nais/ds-svelte-community';
	import type { PageProps } from './$types';
	import TeamsTable, { type TeamsData } from '../TeamsTable.svelte';
	import type { AllTeams$result } from '$houdini';
	import Tab from '$lib/ui/Tab.svelte';
	import Tabs from '$lib/ui/Tabs.svelte';

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
	<div class="content-wrapper">
		<div class="header">
			<Heading level="1" size="large">Teamoversikt</Heading>
			{#if canCreateTeam}
				<Button as="a" size="medium" href="/team/create" variant="primary">Opprett team</Button>
			{/if}
		</div>
		{#if $AllTeams.data}
			<div class="container">
				<div>
					<Tabs>
						<Tab href="/" active={false} title="Mine team ({userTeamsCount})" />
						<Tab href="/teams" active={true} title="Alle teams ({allTeamsCount})" />
					</Tabs>

					<TeamsTable
						defaultSelected={['name', 'memberCount', 'manager']}
						teamsData={$AllTeams.data.teams.nodes.map(transformTeamData)}
					/>
				</div>
			</div>
			<Pagination
				page={$AllTeams.data.teams.pageInfo}
				loaders={{
					loadPreviousPage: () => AllTeams.loadPreviousPage(),
					loadNextPage: () => AllTeams.loadNextPage()
				}}
			/>
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
</style>
