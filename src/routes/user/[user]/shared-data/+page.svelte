<script lang="ts">
	import type { PageProps } from './$types';
	import Pagination from '$lib/ui/Pagination.svelte';
	import BucketsTable from './BucketsTable.svelte';

	let { data }: PageProps = $props();

	let { UserSharedBucketAccess } = $derived(data);
</script>

<div class="container">
	<BucketsTable
		bucketsData={$UserSharedBucketAccess.data?.user.sharedBucketsAccess.nodes ?? []}
		defaultSelected={data.bucketTableFields}
	/>
</div>

<Pagination
	page={$UserSharedBucketAccess.data?.user.sharedBucketsAccess.pageInfo}
	loaders={{
		loadPreviousPage: () => UserSharedBucketAccess.loadPreviousPage(),
		loadNextPage: () => UserSharedBucketAccess.loadNextPage()
	}}
/>
