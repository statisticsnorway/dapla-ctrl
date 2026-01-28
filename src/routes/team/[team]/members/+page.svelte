<script lang="ts">
	import GraphErrors from '$lib/ui/GraphErrors.svelte';
	import Pagination from '$lib/ui/Pagination.svelte';
	import type { PageProps } from './$houdini';
	import GroupMembersTable, { type GroupMembersData } from './GroupMembersTable.svelte';
	import type { Groups$result } from '$houdini';

	let { data }: PageProps = $props();
	let { Groups } = $derived(data);

	type GroupMemberNode = Groups$result['team']['members']['nodes'][0];
	function transformGroupMembersData(groupMember: GroupMemberNode): GroupMembersData {
		return {
			name: groupMember.user.name,
			email: groupMember.user.email,
			section: {
				code: groupMember.user.section?.code ?? '',
				name: groupMember.user.section?.name ?? ''
			},
			groups: groupMember.groups.map((group) => ({
				category: group.category,
				suffix: group.suffix ?? null
			}))
		};
	}
</script>

<GraphErrors errors={$Groups.errors} />

<div class="container">
	<GroupMembersTable
		groupMembersData={$Groups.data?.team?.members.nodes.map(transformGroupMembersData) ?? []}
	/>
</div>

<Pagination
	page={$Groups.data?.team?.members.pageInfo}
	loaders={{
		loadPreviousPage: () => Groups.loadPreviousPage(),
		loadNextPage: () => Groups.loadNextPage()
	}}
/>
