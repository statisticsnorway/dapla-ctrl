<script lang="ts">
	import type { PageProps } from './$types';
	import Pagination from '$lib/ui/Pagination.svelte';
	import type { TeamSharedBucketAccessTeams$result } from '$houdini';
	import { BodyShort, CopyButton } from '@nais/ds-svelte-community';
	import DaplaTable from '$lib/ui/DaplaTable.svelte';
	import { capitalizeFirstLetter } from '$lib/utils/formatters';
	import Tab from '$lib/ui/Tab.svelte';
	import Tabs from '$lib/ui/Tabs.svelte';
	import { page } from '$app/state';

	let { params, data }: PageProps = $props();

	let { TeamSharedBucketAccessTeams } = $derived(data);

	type TeamItem = TeamSharedBucketAccessTeams$result['sharedBucket']['teams']['nodes'][0];
</script>

{#snippet nameCell(item: TeamItem)}
	<a href={`/team/${item.slug}/`}>
		<b>{item.displayName}</b>
	</a>
	<br />
	{item.slug}
{/snippet}
{#snippet autonomyCell(item: TeamItem)}
	{#if item.isManaged}
		Managed
	{:else}
		Self-managed
	{/if}
{/snippet}
{#snippet membersCell(item: TeamItem)}
	{item.members.pageInfo.totalCount}
{/snippet}
{#snippet managerCell(item: TeamItem)}
	{#if item.section.manager}
		<a href="/user/{item.section.manager.email}">{item.section.manager.name}</a>
	{:else}
		Mangler seksjonsleder
	{/if}
	<br />
	{item.section.name} ({item.section.code})
{/snippet}
<div class="bucket-info">
	<div>
		<BodyShort
			>{params.bucket}
			<CopyButton
				copyText={params.bucket}
				title={params.bucket}
				iconPosition="right"
				size="xsmall"
			/>
			<br />
			{capitalizeFirstLetter($TeamSharedBucketAccessTeams.data?.sharedBucket.kind ?? '')}
		</BodyShort>
	</div>
</div>

{#if $TeamSharedBucketAccessTeams.data}
	<div class="container">
		<Tabs>
			<Tab
				data-sveltekit-noscroll
				href={`/team/${params.team}/shared-data/${params.bucket}`}
				active={page.url.pathname === `/team/${params.team}/shared-data/${params.bucket}`}
				title="Medlemmer ({$TeamSharedBucketAccessTeams.data.sharedBucket.uniqueUsers.pageInfo
					.totalCount})"
			/>
			<Tab
				data-sveltekit-noscroll
				href={`/team/${params.team}/shared-data/${params.bucket}/teams`}
				active={page.url.pathname === `/team/${params.team}/shared-data/${params.bucket}/teams`}
				title={`Team (${$TeamSharedBucketAccessTeams.data.sharedBucket.teams.pageInfo.totalCount})`}
			/>
		</Tabs>
		<DaplaTable
			data={$TeamSharedBucketAccessTeams.data.sharedBucket.teams.nodes}
			fieldsCookie={{
				path: '/team',
				key: 'sharedBucketUsersFieldsTeams/team'
			}}
			selected={data.bucketTableFields}
			columns={[
				{
					id: 'NAME',
					show: 'ALWAYS',
					name: 'Navn',
					cell: nameCell
				},
				{
					id: 'AUTONOMY',
					show: 'DEFAULT_NO',
					name: 'Autonomitetsnivå',
					cell: autonomyCell
				},
				{
					id: 'MEMBER_COUNT',
					show: 'DEFAULT_YES',
					name: 'Teammedlemmer',
					cell: membersCell
				},
				{
					id: 'MANAGER',
					show: 'DEFAULT_YES',
					name: 'Ansvarlig',
					cell: managerCell
				}
			]}
		/>
	</div>

	<Pagination
		page={$TeamSharedBucketAccessTeams.data?.sharedBucket.teams.pageInfo}
		loaders={{
			loadPreviousPage: () => TeamSharedBucketAccessTeams.loadPreviousPage(),
			loadNextPage: () => TeamSharedBucketAccessTeams.loadNextPage()
		}}
	/>
{/if}

<style>
	.bucket-info {
		display: grid;
		grid-template-columns: 1fr 300px;
		gap: var(--spacing-layout);
		margin-top: calc(-1 * var(--spacing-layout));
		margin-bottom: var(--spacing-layout);
	}
</style>
