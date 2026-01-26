<script lang="ts">
	import Pagination from '$lib/ui/Pagination.svelte';
	import { Button, Heading } from '@nais/ds-svelte-community';
	import Tab from '$lib/ui/Tab.svelte';
	import Tabs from '$lib/ui/Tabs.svelte';
	import TeamsTable, { type TeamsData } from '../../routes/TeamsTable.svelte';
	import { page } from '$app/state';

	interface Props {
		canCreateTeam: boolean;
		userTeamsCount: number;
		allTeamsCount: number;
		teamsData: TeamsData[];
		teamTableDefaultFields: string[];
		loaders: {
			loadNextPage: () => unknown;
			loadPreviousPage: () => unknown;
		};
		pageInfo:
			| {
					readonly hasNextPage: boolean;
					readonly hasPreviousPage: boolean;
					readonly pageStart: number;
					readonly pageEnd: number;
					readonly totalCount: number;
			  }
			| undefined;
	}
	let {
		canCreateTeam,
		userTeamsCount,
		allTeamsCount,
		teamTableDefaultFields,
		teamsData,
		pageInfo,
		loaders
	}: Props = $props();
</script>

<div class="content-wrapper">
	<div class="header">
		<Heading level="1" size="xlarge">Oversikt</Heading>
		{#if canCreateTeam}
			<Button as="a" size="medium" href="/team/create" variant="secondary">+ Opprett team</Button>
		{/if}
	</div>
	<div class="container" data-sveltekit-preload-data="hover">
		<div>
			<Tabs>
				<Tab
					data-sveltekit-noscroll
					href="/"
					active={page.url.pathname === '/'}
					title="Mine team ({userTeamsCount})"
				/>
				<Tab
					data-sveltekit-noscroll
					href="/teams"
					active={page.url.pathname === '/teams'}
					title="Alle teams ({allTeamsCount})"
				/>
			</Tabs>

			<TeamsTable defaultSelected={teamTableDefaultFields} {teamsData} />
		</div>
	</div>
	<Pagination page={pageInfo} {loaders} fetching={!teamsData} />
</div>

<style>
	.content-wrapper {
		background: var(--ax-bg-default);
		position: relative;
		margin: auto;
		max-width: 1432px;
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
