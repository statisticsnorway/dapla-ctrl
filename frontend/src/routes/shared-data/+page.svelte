<script lang="ts">
	import type { PageProps } from './$types';
	import Pagination from '$lib/ui/Pagination.svelte';
	import { Heading, BodyShort } from '@nais/ds-svelte-community';
	import { SharedBucketOrderField, type AllSharedData$result } from '$houdini';
	import DaplaTable from '$lib/ui/DaplaTable.svelte';
	import { capitalizeFirstLetter } from '$lib/utils/formatters';

	let { data }: PageProps = $props();

	let { AllSharedData } = $derived(data);

	type BucketNode = AllSharedData$result['sharedBuckets']['nodes'][0];

	type BucketData = {
		id: string;
		team: {
			slug: string;
			displayName: string;
		};
		name: string;
		shortName: string;
		type: string;
		env: string;
		teamCount: number;
		userCount: number;
		sectionCode: string;
	};

	function transformBucketdata(bucketNode: BucketNode): BucketData {
		return {
			id: bucketNode.id,
			name: bucketNode.name,
			team: bucketNode.team,
			shortName: bucketNode.shortName,
			env: bucketNode.env,
			type: bucketNode.kind,
			teamCount: bucketNode.teams.pageInfo.totalCount,
			userCount: bucketNode.uniqueUsers.pageInfo.totalCount,
			sectionCode: bucketNode.team.section.code
		};
	}
</script>

{#snippet nameCell(bucket: BucketData)}
	<a href={`/team/${bucket.team.slug}/shared-data/${bucket.name}`}>
		<b>{bucket.shortName}</b>
	</a>
	<br />
	{bucket.name}
{/snippet}
{#snippet typeCell(bucket: BucketData)}
	{capitalizeFirstLetter(bucket.type)}{/snippet}
{#snippet teamCell(bucket: BucketData)}<a href={`/team/${bucket.team.slug}/`}>
		<b>{bucket.team.displayName}</b>
	</a>
	<br />
	{bucket.team.slug}{/snippet}
{#snippet sectionCell(bucket: BucketData)}{bucket.sectionCode}{/snippet}
{#snippet envCell(bucket: BucketData)}{bucket.env}{/snippet}
{#snippet teamsCell(bucket: BucketData)}{bucket.teamCount}{/snippet}
{#snippet usersCell(bucket: BucketData)}{bucket.userCount}{/snippet}

<svelte:head><title>Datadeling - Dapla Ctrl</title></svelte:head>

<div class="page">
	<div class="header">
		<Heading level="1" size="xlarge">Datadeling</Heading>
		<div class="description">
			<BodyShort textColor="subtle" size="medium">
				Oversikt over alle deltbøtter på Dapla.
			</BodyShort>
		</div>
	</div>
	<div class="container">
		<DaplaTable
			data={$AllSharedData.data?.sharedBuckets.nodes.map(transformBucketdata) ?? []}
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
					id: 'TEAM',
					name: 'Team',
					show: 'DEFAULT_YES',
					cell: teamCell,
					sortKey: SharedBucketOrderField.TEAM
				},
				{
					id: 'SECTION',
					name: 'Seksjon',
					align: 'right',
					show: 'DEFAULT_YES',
					cell: sectionCell
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
					align: 'right',
					show: 'DEFAULT_YES',
					cell: teamsCell
				},
				{
					id: 'USER_COUNT',
					name: 'Antall personer',
					align: 'right',
					show: 'DEFAULT_YES',
					cell: usersCell
				}
			]}
		/>
	</div>

	<Pagination
		page={$AllSharedData.data?.sharedBuckets.pageInfo}
		loaders={{
			loadPreviousPage: () => AllSharedData.loadPreviousPage(),
			loadNextPage: () => AllSharedData.loadNextPage()
		}}
	/>
</div>

<style>
	.description {
		margin-top: var(--ax-space-8);
		margin-bottom: var(--ax-space-16);
	}
</style>
