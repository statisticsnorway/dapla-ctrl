<script lang="ts">
	import type { PageProps } from './$types';
	import Pagination from '$lib/ui/Pagination.svelte';
	import type { ConsumesSharedData$result } from '$houdini';
	import DaplaTable from '$lib/ui/DaplaTable.svelte';
	import { capitalizeFirstLetter } from '$lib/utils/formatters';
	import Tab from '$lib/ui/Tab.svelte';
	import Tabs from '$lib/ui/Tabs.svelte';
	import { page } from '$app/state';

	let { data, params }: PageProps = $props();

	let { ConsumesSharedData, teamSlug } = $derived(data);

	type BucketItem = ConsumesSharedData$result['team']['sharedBucketsAccess']['nodes'][0];
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

{#if $ConsumesSharedData.data}
	<div class="container">
		<Tabs>
			<Tab
				data-sveltekit-noscroll
				href={`/team/${params.team}/shared-data`}
				active={page.url.pathname === `/team/${params.team}/shared-data`}
				title="Deler ({$ConsumesSharedData.data.team.sharedBuckets.pageInfo.totalCount})"
			/>
			<Tab
				data-sveltekit-noscroll
				href={`/team/${params.team}/shared-data/consumes`}
				active={page.url.pathname === `/team/${params.team}/shared-data/consumes`}
				title={`Konsumerer (${$ConsumesSharedData.data.team.sharedBucketsAccess.pageInfo.totalCount})`}
			/>
		</Tabs>
		<DaplaTable
			data={$ConsumesSharedData.data.team.sharedBucketsAccess.nodes}
			fieldsCookie={{ path: '/team', key: 'consumesSharedBucketsFields/team' }}
			selected={data.bucketTableFields}
			columns={[
				{
					id: 'NAME',
					name: 'Navn',
					show: 'ALWAYS',
					cell: nameCell
				},
				{
					id: 'TYPE',
					name: 'Type',
					show: 'DEFAULT_YES',
					cell: typeCell
				},
				{
					id: 'ENV',
					name: 'Miljø',
					show: 'DEFAULT_YES',
					cell: envCell
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
		page={$ConsumesSharedData.data?.team.sharedBucketsAccess.pageInfo}
		loaders={{
			loadPreviousPage: () => ConsumesSharedData.loadPreviousPage(),
			loadNextPage: () => ConsumesSharedData.loadNextPage()
		}}
	/>
{/if}
