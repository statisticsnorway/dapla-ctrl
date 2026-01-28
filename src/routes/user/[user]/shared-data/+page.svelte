<script lang="ts">
	import type { PageProps } from './$types';
	import Pagination from '$lib/ui/Pagination.svelte';
	import DaplaTable from '$lib/ui/DaplaTable.svelte';
	import type { UserSharedBucketAccess$result } from '$houdini';
	import { capitalizeFirstLetter } from '$lib/utils/formatters';
	import { BodyShort } from '@nais/ds-svelte-community';

	let { data }: PageProps = $props();

	let { UserSharedBucketAccess } = $derived(data);

	type BucketItem = UserSharedBucketAccess$result['user']['sharedBucketsAccess']['nodes'][0];
</script>

<div class="description">
	<BodyShort textColor="subtle" size="medium">
		Oversikt over hvilke delt-bøtter medlemmet har tilgang til.
	</BodyShort>
</div>

{#snippet nameCell(bucket: BucketItem)}
	<a href={`/team/${bucket.team.slug}/shared-data/${bucket.name}`}><b>{bucket.shortName}</b></a>
	<br />
	{bucket.name}
{/snippet}

{#snippet typeCell(bucket: BucketItem)}
	{capitalizeFirstLetter(bucket.kind)}
{/snippet}

{#snippet envCell(bucket: BucketItem)}
	{bucket.env}
{/snippet}

{#snippet teamCell(bucket: BucketItem)}
	<a href={`/team/${bucket.team.slug}`}><b>{bucket.team.displayName}</b></a>
	<br />
	{bucket.team.slug}
{/snippet}

{#if $UserSharedBucketAccess.data}
	<div class="container">
		<DaplaTable
			data={$UserSharedBucketAccess.data.user.sharedBucketsAccess.nodes ?? []}
			fieldsCookie={{
				path: '/user',
				key: 'sharedBucketsTableFields/user'
			}}
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
					id: 'TEAM',
					name: 'Team',
					show: 'DEFAULT_YES',
					cell: teamCell
				}
			]}
		/>
	</div>

	<Pagination
		page={$UserSharedBucketAccess.data?.user.sharedBucketsAccess.pageInfo}
		loaders={{
			loadPreviousPage: () => UserSharedBucketAccess.loadPreviousPage(),
			loadNextPage: () => UserSharedBucketAccess.loadNextPage()
		}}
	/>
{/if}

<style>
	.description {
		margin-top: calc(-1 * var(--spacing-layout));
		margin-bottom: var(--ax-space-16);
	}
</style>
