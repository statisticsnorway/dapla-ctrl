<script lang="ts">
	import GraphErrors from '$lib/ui/GraphErrors.svelte';
	import Pagination from '$lib/ui/Pagination.svelte';
	import type { PageProps } from './$houdini';
	import { UserOrderField, type Groups$result } from '$houdini';
	import DaplaTable from '$lib/ui/DaplaTable.svelte';
	import { BodyShort, Button } from '@nais/ds-svelte-community';
	import { PencilIcon, PlusIcon } from '@nais/ds-svelte-community/icons';
	import AddMember from './AddMember.svelte';
	import EditMember from './EditMember.svelte';

	let { data }: PageProps = $props();
	let { Groups, displayName, UserInfo, teamSlug } = $derived(data);

	// TODO: When releasing to the public, replace this with simply `viewerCanManageMembers` in the
	// destructure on line 13
	let viewerCanManageMembers = $derived(
		$UserInfo.data?.me.__typename === 'User' && $UserInfo.data.me.isAdmin
	);

	let groups = $derived($Groups.data?.team.groups.nodes);

	let addMemberOpen = $state(false);

	let modifyMember: {
		open: boolean;
		user: { name: string; email: string };
	} = $state({
		open: false,
		user: { name: '', email: '' }
	});

	let currentGroups: string[] = $state([]);

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
			name: string;
			category: string;
			suffix: string | null;
		}[];
	};

	const refetch = () => {
		Groups.fetch({
			policy: 'CacheAndNetwork'
		});
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
				name: group.name,
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
	<div style="display: flex; flex-direction: row;">
		<div>
			{groupMember.groups
				.map((g) => g.name.substring(teamSlug.length + 1))
				.toSorted()
				.join(', ')}
		</div>
		{#if viewerCanManageMembers}
			<div style="margin-left: auto; display: flex;">
				<Button
					size="small"
					variant="tertiary"
					onclick={() => {
						currentGroups = groupMember.groups.map((g) => g.name);
						modifyMember = {
							open: true,
							user: {
								name: groupMember.name,
								email: groupMember.email
							}
						};
					}}
					icon={PencilIcon}
				></Button>
			</div>
		{/if}
	</div>
{/snippet}

<GraphErrors errors={$Groups.errors} />

<div class="description">
	<BodyShort textColor="subtle" size="medium">
		Oversikt over {displayName} sine medlemmer.
	</BodyShort>
</div>

{#if $Groups.data}
	<div class="container">
		{#if viewerCanManageMembers}
			<div class="button">
				<Button
					size="small"
					onclick={() => {
						addMemberOpen = !addMemberOpen;
					}}
					icon={PlusIcon}>Legg til medlem</Button
				>
			</div>
		{/if}

		<DaplaTable
			data={$Groups.data?.team?.members.nodes.map(transformGroupMembersData) ?? []}
			fieldsCookie={{
				path: '/team',
				key: 'teamMembersTableFields'
			}}
			selected={data.groupMemberTableFields}
			columns={[
				{
					id: 'NAME',
					name: 'Navn',
					show: 'ALWAYS',
					cell: nameCell,
					sortKey: UserOrderField.NAME
				},
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

	{#if groups}
		<AddMember bind:open={addMemberOpen} {groups} team={data.teamSlug} oncreated={refetch} />
		{#if modifyMember.open}
			<EditMember
				bind:open={modifyMember.open}
				groups={groups.map((g) => g.name)}
				{currentGroups}
				user={modifyMember.user}
				team={data.teamSlug}
				oncreated={refetch}
			/>
		{/if}
	{/if}
{/if}

<style>
	.button {
		display: flex;
		justify-content: flex-end;
		margin-bottom: var(--ax-space-24, --a-spacing-6);
	}
	.description {
		margin-top: calc(-1 * var(--spacing-layout));
		margin-bottom: var(--ax-space-16);
	}
</style>
