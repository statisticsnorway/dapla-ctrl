<script lang="ts">
	import type { PageProps } from './$types';
	import Pagination from '$lib/ui/Pagination.svelte';
	import type { AllSharedData$result } from '$houdini';
	import BucketsTable, { type BucketData } from './BucketsTable.svelte';
	import { Heading } from '@nais/ds-svelte-community';

	let { data }: PageProps = $props();

	let { AllSharedData } = $derived(data);

	type BucketNode = AllSharedData$result['sharedBuckets']['nodes'][0];
	function transformBucketdata(bucketNode: BucketNode): BucketData {
		return {
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

<svelte:head><title>Datadeling - Dapla Ctrl</title></svelte:head>

<div class="page">
	<div class="content-wrapper">
		<div class="header">
			<Heading level="1" size="xlarge">Datadeling</Heading>
		</div>
		<div class="container">
			<BucketsTable
				bucketsData={$AllSharedData.data?.sharedBuckets.nodes.map(transformBucketdata) ?? []}
				defaultSelected={data.bucketTableFields}
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
</div>

<style>
	.page {
		margin-inline: var(--margin-default);
	}
</style>
