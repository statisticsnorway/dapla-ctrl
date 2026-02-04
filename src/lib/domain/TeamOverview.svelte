<script lang="ts">
	import Pagination from '$lib/ui/Pagination.svelte';
	import { BodyShort, Button, Heading } from '@nais/ds-svelte-community';
	import Tab from '$lib/ui/Tab.svelte';
	import Tabs from '$lib/ui/Tabs.svelte';
	import { page } from '$app/state';
	import DaplaTable from '$lib/ui/DaplaTable.svelte';
	import { TeamOrderField } from '$houdini';

	type User = {
		name: string;
		email: string;
	};
	export type TeamsData = {
		id: string;
		slug: string;
		displayName: string;
		memberCount: number;
		manager: User;
		isManaged: boolean;
		section: {
			code: string;
			name: string;
		};
		userGroups?: string[];
	};
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

	const isAllTeamsPage = $derived(page.url.pathname === '/teams');

	const description = $derived(
		isAllTeamsPage ? 'Oversikt over alle team.' : 'Oversikt over alle team du er medlem av.'
	);

	const columns = $derived(
		[
			{
				id: 'NAME',
				name: 'Navn',
				show: 'ALWAYS',
				cell: nameCell,
				sortKey: TeamOrderField.SLUG
			} as const,
			{
				id: 'AUTONOMY',
				name: 'Autonomitetsnivå',
				show: 'DEFAULT_NO',
				cell: autonomyCell
			} as const,
			isAllTeamsPage
				? undefined
				: ({
						id: 'GROUPS',
						name: 'Mine roller',
						show: 'DEFAULT_YES',
						cell: groupsCell
					} as const),
			{
				id: 'MEMBER_COUNT',
				name: 'Teammedlemmer',
				align: 'right',
				show: 'DEFAULT_YES',
				cell: membersCell
			} as const,
			{
				id: 'MANAGER',
				name: 'Ansvarlig',
				show: 'DEFAULT_YES',
				cell: managerCell,
				sortKey: TeamOrderField.SECTION_CODE
			} as const
		].filter((c) => c !== undefined)
	);
</script>

{#snippet nameCell(team: TeamsData)}
	<a href={`/team/${team.slug}/`}>
		<b>{team.displayName}</b>
	</a>
	<br />
	{team.slug}
{/snippet}
{#snippet autonomyCell(team: TeamsData)}
	{#if team.isManaged}
		Managed
	{:else}
		Self-managed
	{/if}
{/snippet}
{#snippet groupsCell(team: TeamsData)}
	{team.userGroups
		?.map((g) => g.substring(team.slug.length + 1))
		.toSorted()
		.join(', ') ?? []}
{/snippet}
{#snippet membersCell(team: TeamsData)}
	{team.memberCount}
{/snippet}
{#snippet managerCell(team: TeamsData)}
	{#if team.manager.email !== ''}
		<a href="/member/{team.manager.email}">{team.manager.name}</a>
	{:else}
		{team.manager.name}
	{/if}
	<br />
	{team.section.name} ({team.section.code})
{/snippet}

<div class="content-wrapper">
	<div class="header">
		<div>
			<Heading level="1" size="xlarge">Team</Heading>
			<div class="description">
				<BodyShort textColor="subtle" size="medium">{description}</BodyShort>
			</div>
		</div>
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
					title="Alle team ({allTeamsCount})"
				/>
			</Tabs>

			<DaplaTable data={teamsData} selected={teamTableDefaultFields} {columns} />
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

	.description {
		margin-top: var(--ax-space-4);
	}

	.container {
		margin-top: var(--spacing-layout);
		display: flex;
		flex-direction: column;
		gap: var(--spacing-layout);
	}
</style>
