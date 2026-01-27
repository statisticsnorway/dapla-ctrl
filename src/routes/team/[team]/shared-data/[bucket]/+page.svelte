<script lang="ts">
	import type { PageProps } from './$types';
	import Pagination from '$lib/ui/Pagination.svelte';
	import type { TeamSharedBucketAccess$result } from '$houdini';
	import { BodyShort, CopyButton } from '@nais/ds-svelte-community';
	import DaplaTable from '$lib/ui/DaplaTable.svelte';

	let { params, data }: PageProps = $props();

	let { TeamSharedBucketAccess } = $derived(data);

	type TeamMemberItem = TeamSharedBucketAccess$result['sharedBucket']['users']['nodes'][0] & {
		id: string;
	};
</script>

{#snippet nameCell(teamMember: TeamMemberItem)}
	<a href={`/user/${teamMember.user.email}/shared-data`}>
		<b>{teamMember.user.name}</b>
	</a>
	<br />
	{teamMember.user.email}
{/snippet}
{#snippet sectionCell(teamMember: TeamMemberItem)}
	{#if teamMember.user.section}
		{#if teamMember.user.section.manager}
			<a href="/user/{teamMember.user.section.manager.email}"
				>{teamMember.user.section.manager.name}</a
			>
		{:else}
			Mangler seksjonsleder
		{/if}
		<br />
		{teamMember.user.section.name} ({teamMember.user.section.code})
	{/if}
{/snippet}
{#snippet teamCell(teamMember: TeamMemberItem)}
	<a href={`/team/${teamMember.team.slug}/`}>
		<b>{teamMember.team.displayName}</b>
	</a>
	<br />
	{teamMember.team.slug}
{/snippet}
{#snippet groupCell(teamMember: TeamMemberItem)}
	{teamMember.groups.map((g) => g.name.slice(teamMember.team.slug.length + 1)).join(', ')}
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
		</BodyShort>
	</div>
</div>

{#if $TeamSharedBucketAccess.data}
	<div class="container">
		<DaplaTable
			data={$TeamSharedBucketAccess.data.sharedBucket.users.nodes.map((tm) => {
				return { id: `${tm.team.id}:${tm.user.id}`, ...tm };
			})}
			fieldsCookie={{
				path: '/team',
				key: 'sharedBucketUsersFields/team'
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
					id: 'SECTION',
					name: 'Ansvarlig',
					show: 'DEFAULT_YES',
					cell: sectionCell
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
		page={$TeamSharedBucketAccess.data?.sharedBucket.users.pageInfo}
		loaders={{
			loadPreviousPage: () => TeamSharedBucketAccess.loadPreviousPage(),
			loadNextPage: () => TeamSharedBucketAccess.loadNextPage()
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
