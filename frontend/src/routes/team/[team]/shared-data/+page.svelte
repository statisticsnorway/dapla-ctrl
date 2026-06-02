<script lang="ts">
	import type { PageProps } from './$types';
	import Pagination from '$lib/ui/Pagination.svelte';
	import { SharedBucketOrderField, type SharedData$result } from '$houdini';
	import DaplaTable from '$lib/ui/DaplaTable.svelte';
	import { capitalizeFirstLetter } from '$lib/utils/formatters';
	import Tab from '$lib/ui/Tab.svelte';
	import Tabs from '$lib/ui/Tabs.svelte';
	import { page } from '$app/state';
	import { BodyShort } from '@nais/ds-svelte-community';

	let { data, params }: PageProps = $props();

	let { SharedData, teamSlug, displayName } = $derived(data);

	type BucketItem = SharedData$result['team']['sharedBuckets']['nodes'][0];
</script>

{#snippet nameCell(bucket: BucketItem)}
	<a href={`/team/${teamSlug}/shared-data/${bucket.name}`}><b>{bucket.shortName}</b></a>
	<br />
	{bucket.name}
{/snippet}
{#snippet typeCell(bucket: BucketItem)}
	{capitalizeFirstLetter(bucket.kind)}
{/snippet}
{#snippet envCell(bucket: BucketItem)}
	{bucket.env}
{/snippet}
{#snippet teamsCell(bucket: BucketItem)}
	{bucket.teams.pageInfo.totalCount}
{/snippet}
{#snippet usersCell(bucket: BucketItem)}
	{bucket.uniqueUsers.pageInfo.totalCount}
{/snippet}

<div class="description">
	<BodyShort textColor="subtle" size="medium">
		Oversikt over hvilke deltbøtter {displayName} deler.
	</BodyShort>
</div>

{#if $SharedData.data}
	<div class="container">
		<Tabs>
			<Tab
				data-sveltekit-noscroll
				href="/team/${params.team}/shared-data"
				active={page.url.pathname === `/team/${params.team}/shared-data`}
				title={`Deler (${$SharedData.data.team.sharedBuckets.pageInfo.totalCount})`}
			/>
			<Tab
				data-sveltekit-noscroll
				href={`/team/${params.team}/shared-data/consumes`}
				active={page.url.pathname === `/team/${params.team}/shared-data/consumes`}
				title={`Konsumerer (${$SharedData.data.team.sharedBucketsAccess.pageInfo.totalCount})`}
			/>
		</Tabs>
		<DaplaTable
			data={$SharedData.data.team.sharedBuckets.nodes}
			fieldsCookie={{ path: '/team', key: 'sharedBucketsFields/team' }}
			selected={data.bucketTableFields}
			columns={[
				{
					id: 'NAME',
					name: 'Navn',
					show: 'ALWAYS',
					cell: nameCell,
					sortKey: SharedBucketOrderField.SHORT_NAME
				},
				{
					id: 'TYPE',
					name: 'Type',
					show: 'DEFAULT_YES',
					cell: typeCell,
					sortKey: SharedBucketOrderField.KIND
				},
				{
					id: 'ENV',
					name: 'Miljø',
					show: 'DEFAULT_YES',
					cell: envCell,
					sortKey: SharedBucketOrderField.ENV
				},
				{
					id: 'TEAM_COUNT',
					name: 'Antall team',
					show: 'DEFAULT_YES',
					cell: teamsCell
				},
				{
					id: 'USER_COUNT',
					name: 'Antall personer',
					show: 'DEFAULT_YES',
					cell: usersCell
				}
			]}
		/>
	</div>

	<Pagination
		page={$SharedData.data?.team.sharedBuckets.pageInfo}
		loaders={{
			loadPreviousPage: () => SharedData.loadPreviousPage(),
			loadNextPage: () => SharedData.loadNextPage()
		}}
	/>
{/if}

<style>
	.description {
		margin-top: calc(-1 * var(--spacing-layout));
		margin-bottom: var(--ax-space-16);
	}
</style>
