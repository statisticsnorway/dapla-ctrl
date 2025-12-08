<script lang="ts">
	import Pagination from '$lib/ui/Pagination.svelte';
	import { Button, Heading } from '@nais/ds-svelte-community';
	import type { PageProps } from './$types';
	import TeamsTable, { type TeamsData } from '../TeamsTable.svelte';
	import type { AllTeams$result } from '$houdini';

	let { data }: PageProps = $props();

	let { AllTeams } = $derived(data);

	type TeamNode = AllTeams$result['teams']['nodes'][0];
	type GroupNode = TeamNode['groups']['nodes'][0];
	type GroupUserNode = GroupNode['members']['nodes'][0];

	let allTeamsCount = $derived($AllTeams.data?.teams?.nodes.length || 0);
	function transformTeamData(teamNode: TeamNode): TeamsData {
		const team = teamNode;

		const allManagers = team.groups.nodes
			.filter((group: GroupNode) => group.category === 'managers' && !group.suffix)
			.flatMap((managerGroup: GroupNode) =>
				managerGroup.members.nodes.map((member: GroupUserNode) => member.user)
			);
		const uniqueManagers = Array.from(
			new Map(allManagers.map((user: GroupUserNode['user']) => [user.email, user])).values()
		);

		return {
			slug: team.slug,
			memberCount: team.members.pageInfo.totalCount,
			managers: uniqueManagers
		};
	}
</script>

<svelte:head><title>Alle team - Dapla Ctrl</title></svelte:head>

<div class="page">
	<div class="content-wrapper">
		<div class="header">
			<Heading level="1" size="large">Alle team</Heading>
			<Button as="a" size="medium" href="/team/create" variant="primary">Opprett team</Button>
		</div>
		{#if $AllTeams.data}
			<div class="container">
				<div>
					<div class="section-header">
						<Heading level="2" spacing>Alle team ({allTeamsCount})</Heading>
						<Button as="a" size="small" href="/" variant="secondary">Se mine team</Button>
					</div>

					<TeamsTable teamsData={$AllTeams.data.teams.nodes.map(transformTeamData)} />
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

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--ax-space-16);
	}
</style>
