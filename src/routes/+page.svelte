<script lang="ts">
	import Pagination from '$lib/ui/Pagination.svelte';
	import { Button, Heading } from '@nais/ds-svelte-community';
	import type { PageProps } from './$types';
	import TeamsTable from './TeamsTable.svelte';

	let { data }: PageProps = $props();

	let { UserTeams, UserInfo } = $derived(data);

	let userTeams = $derived(
		$UserTeams.data?.me.__typename == 'User' && $UserTeams.data.me.teams?.nodes.length
	);

	let name = $derived($UserInfo.data?.me.__typename == 'User' ? $UserInfo.data.me.name : '');
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
				<TeamsTable
					teamsData={$UserTeams.data.me.teams.nodes.map((node) => {
						return {
							slug: node.team.slug,
							memberCount: node.team.members.pageInfo.totalCount,
							managers: node.team.groups.nodes
								.filter((group) => group.category === 'managers')
								.flatMap((managerGroup) => managerGroup.members.nodes.map((member) => member.user))
						};
					})}
				/>
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
</style>
