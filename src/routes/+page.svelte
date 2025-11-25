<script lang="ts">
	import List from '$lib/components/list/List.svelte';
	import TeamsTable from './TeamsTable.svelte';
	import TeamListItem from '$lib/components/list/TeamListItem.svelte';
	import Pagination from '$lib/Pagination.svelte';
	import { BodyLong, Button, Heading } from '@nais/ds-svelte-community';
	import type { PageProps } from './$houdini';

	let { data }: PageProps = $props();

	let UserTeams = $derived(data.UserTeams);
	let tenantName = $derived(data.tenantName);

	let userTeams = $derived(
		$UserTeams.data?.me.__typename == 'User' && $UserTeams.data?.me.teams?.nodes.length
	);
</script>

<svelte:head><title>Dapla Ctrl</title></svelte:head>

<div class="page">
	<div class="content-wrapper">
		<div class="header">
			<Heading level="1" size="large">Teamoversikt</Heading>
			<Button as="a" size="medium" href="/team/create" variant="primary">Opprett team</Button>
		</div>
		{#if $UserTeams.data}
			{#if $UserTeams.data.me.__typename == 'User'}
			    <TeamsTable teamsData={$UserTeams.data.me.teams.nodes.map(node => {return {
						slug: node.team.slug,
						memberCount: node.team.members.pageInfo.totalCount,
						managers: node.team.groups.nodes.filter(group => group.category === "managers").flatMap(managerGroup => managerGroup.members.nodes.map(member => member.user))
			    }})}  />
				<List>
					{#each $UserTeams.data.me.teams.nodes as node (node.team.id)}
						<TeamListItem team={node.team} />
					{:else}
						<BodyLong>
							You don't seem to belong to any teams at the moment. You can create a new team or
							search for the team you'd like to join. Once you find it, locate one of the owners in
							the members list on the team page to request membership.
						</BodyLong>
					{/each}
				</List>
				<Pagination
					page={$UserTeams.data.me.teams.pageInfo}
					loaders={{
						loadPreviousPage: UserTeams.loadPreviousPage,
						loadNextPage: UserTeams.loadNextPage
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
		background: var(--ax-bg-default, --a-surface-default);
		position: relative;
		top: -40px;
		padding: var(--ax-space-24, --a-spacing-6);
		border-radius: 12px;
		max-width: 900px;
		margin-inline: auto;
	}

	.header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--ax-space-16, --a-spacing-4);
	}
</style>
