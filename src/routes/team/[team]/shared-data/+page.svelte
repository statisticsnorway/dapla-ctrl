<script lang="ts">
	import type { PageProps } from './$types';
	import Pagination from '$lib/ui/Pagination.svelte';
	import type { SharedData$result } from '$houdini';
	import BucketsTable, { type BucketData } from './BucketsTable.svelte';

	let { data }: PageProps = $props();

	let { SharedData, teamSlug } = $derived(data);

	type BucketNode = SharedData$result['team']['sharedBuckets']['nodes'][0];
	function transformBucketdata(bucketNode: BucketNode): BucketData {
		return {
			name: bucketNode.name,
			shortName: bucketNode.shortName,
			env: bucketNode.env,
			type: bucketNode.kind,
			teamCount: bucketNode.teams.pageInfo.totalCount,
			userCount: bucketNode.uniqueUsers.pageInfo.totalCount
		};
	}
</script>

<div class="container">
	<BucketsTable
		bucketsData={$SharedData.data?.team.sharedBuckets.nodes.map(transformBucketdata) ?? []}
		defaultSelected={data.bucketTableFields}
		{teamSlug}
	/>
</div>

<Pagination
	page={$SharedData.data?.team.sharedBuckets.pageInfo}
	loaders={{
		loadPreviousPage: () => SharedData.loadPreviousPage(),
		loadNextPage: () => SharedData.loadNextPage()
	}}
/>
