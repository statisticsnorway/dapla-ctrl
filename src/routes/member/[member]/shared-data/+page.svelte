<script lang="ts">
	import type { PageProps } from './$types';
	import Pagination from '$lib/ui/Pagination.svelte';
	import DaplaTable from '$lib/ui/DaplaTable.svelte';
	import type { UserSharedBucketAccess$result } from '$houdini';
	import { capitalizeFirstLetter } from '$lib/utils/formatters';
	import { BodyShort } from '@nais/ds-svelte-community';

	let { data }: PageProps = $props();

	let { UserSharedBucketAccess, userDisplayName } = $derived(data);

	type UserAccessNode = UserSharedBucketAccess$result['user']['sharedBucketsAccess']['nodes'][0];

	type BucketAccessItem = UserAccessNode & { id: string };

	const addIdToItems = (items: UserAccessNode[]): BucketAccessItem[] => {
		return items.map((i) => {
			return { id: `${i.bucket.id}_${i.team.id}`, ...i };
		});
	};
</script>

{#snippet nameCell(item: BucketAccessItem)}
	<a href={`/team/${item.team.slug}/shared-data/${item.bucket.name}`}
		><b>{item.bucket.shortName}</b></a
	>
	<br />
	{item.bucket.name}
{/snippet}

{#snippet typeCell(item: BucketAccessItem)}
	{capitalizeFirstLetter(item.bucket.kind)}
{/snippet}

{#snippet envCell(item: BucketAccessItem)}
	{item.bucket.env}
{/snippet}

{#snippet teamCell(item: BucketAccessItem)}
	<a href={`/team/${item.team.slug}`}><b>{item.team.displayName}</b></a>
	<br />
	{item.team.slug}
{/snippet}

{#snippet groupCell(item: BucketAccessItem)}
	{item.groups.map((g) => g.name.slice(item.team.slug.length + 1)).join(', ')}
{/snippet}

<div class="description">
	<BodyShort textColor="subtle" size="medium">
		Oversikt over hvilke deltbøtter {userDisplayName} har tilgang til.
	</BodyShort>
</div>

{#if $UserSharedBucketAccess.data}
	<div class="container">
		<DaplaTable
			data={addIdToItems($UserSharedBucketAccess.data.user.sharedBucketsAccess.nodes ?? [])}
			fieldsCookie={{
				path: '/member',
				key: 'sharedBucketsTableFields/member'
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
					name: 'Tilgang via',
					show: 'DEFAULT_YES',
					cell: [
						{ id: 'team', snippet: teamCell },
						{ id: 'groups', snippet: groupCell }
					]
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
