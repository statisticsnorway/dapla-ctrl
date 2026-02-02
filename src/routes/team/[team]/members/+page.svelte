<script lang="ts">
	import GraphErrors from '$lib/ui/GraphErrors.svelte';
	import Pagination from '$lib/ui/Pagination.svelte';
	import type { PageProps } from './$houdini';
	import { UserOrderField, type Groups$result } from '$houdini';
	import DaplaTable from '$lib/ui/DaplaTable.svelte';
	import { BodyShort } from '@nais/ds-svelte-community';

	let { data }: PageProps = $props();
	let { Groups, displayName } = $derived(data);

	type GroupMemberNode = Groups$result['team']['members']['nodes'][0];

	type GroupMembersData = {
		id: string;
		name: string;
		email: string;
		section: {
			code: string;
			name: string;
		};
		groups: {
			category: string;
			suffix: string | null;
		}[];
	};

	function transformGroupMembersData(groupMemberNode: GroupMemberNode): GroupMembersData {
		return {
			id: groupMemberNode.user.id,
			name: groupMemberNode.user.name,
			email: groupMemberNode.user.email,
			section: {
				code: groupMemberNode.user.section?.code ?? '',
				name: groupMemberNode.user.section?.name ?? ''
			},
			groups: groupMemberNode.groups.map((group) => ({
				category: group.category,
				suffix: group.suffix ?? null
			}))
		};
	}
</script>

{#snippet nameCell(groupMember: GroupMembersData)}
	<a href={`/member/${groupMember.email}`}>
		<b>{groupMember.name}</b>
	</a>
	<br />
	{groupMember.email}
	<br />
	{groupMember.section.name} ({groupMember.section.code})
{/snippet}
{#snippet groupsCell(groupMember: GroupMembersData)}
	{groupMember.groups
		.map((g) => (g.suffix && g.suffix !== '' ? `${g.category}-${g.suffix}` : g.category))
		.toSorted()
		.join(', ')}
{/snippet}

<GraphErrors errors={$Groups.errors} />

<div class="description">
	<BodyShort textColor="subtle" size="medium">
		Oversikt over {displayName} sine medlemmer.
	</BodyShort>
</div>

{#if $Groups.data}
	<div class="container">
		<DaplaTable
			data={$Groups.data?.team?.members.nodes.map(transformGroupMembersData) ?? []}
			fieldsCookie={{
				path: '/team',
				key: 'teamMembersTableFields'
			}}
			selected={data.groupMemberTableFields}
			columns={[
				{ id: 'NAME', name: 'Navn', show: 'ALWAYS', cell: nameCell, sortKey: UserOrderField.NAME },
				{ id: 'GROUPS', name: 'Grupper', show: 'DEFAULT_YES', cell: groupsCell }
			]}
		/>
	</div>
	<Pagination
		page={$Groups.data?.team?.members.pageInfo}
		loaders={{
			loadPreviousPage: () => Groups.loadPreviousPage(),
			loadNextPage: () => Groups.loadNextPage()
		}}
	/>
{/if}

<style>
	.description {
		margin-top: calc(-1 * var(--spacing-layout));
		margin-bottom: var(--ax-space-16);
	}
</style>
