<script lang="ts">
	import type { PageProps } from './$types';
	import Pagination from '$lib/ui/Pagination.svelte';
	import type { SharedData$result } from '$houdini';
	import DaplaTable from '$lib/ui/DaplaTable.svelte';
	import { capitalizeFirstLetter } from '$lib/utils/formatters';

	let { data }: PageProps = $props();

	let { SharedData, teamSlug } = $derived(data);

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

{#if $SharedData.data}
	<div class="container">
		<DaplaTable
			data={$SharedData.data.team.sharedBuckets.nodes}
			fieldsCookie={{ path: '/team', key: 'sharedBucketsFields/team' }}
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
		page={$SharedData.data?.team.sharedBuckets.pageInfo}
		loaders={{
			loadPreviousPage: () => SharedData.loadPreviousPage(),
			loadNextPage: () => SharedData.loadNextPage()
		}}
	/>
{/if}
