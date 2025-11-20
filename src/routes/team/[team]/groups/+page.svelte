<script lang="ts">
	import { graphql, GroupMemberOrderField } from '$houdini';
	import Confirm from '$lib/components/Confirm.svelte';
	import List from '$lib/components/list/List.svelte';
	import ListItem from '$lib/components/list/ListItem.svelte';
	import OrderByMenu from '$lib/components/OrderByMenu.svelte';
	import GraphErrors from '$lib/GraphErrors.svelte';
	// import Pagination from '$lib/Pagination.svelte';
	// import { changeParams } from '$lib/utils/searchparams';
	import { BodyShort, Button, Heading } from '@nais/ds-svelte-community';
	import { PlusIcon, TrashIcon } from '@nais/ds-svelte-community/icons';
	import type { PageProps } from './$houdini';
	import AddMember from './AddMember.svelte';
	import CreateGroup from './CreateGroup.svelte';

	let { data }: PageProps = $props();
	let { Groups, UserInfo, viewerIsOwner } = $derived(data);
	let team = $derived($Groups.data?.team);

	const removeGroupMember = graphql(`
		mutation RemoveGroupMember($input: RemoveGroupMemberInput!) {
			removeGroupMember(input: $input) {
				group {
					name
				}
			}
		}
	`);

	const refetch = () => {
		Groups.fetch({
			policy: 'CacheAndNetwork'
		});
	};

	let addMemberProps: { open: boolean; group: string } = $state({ open: false, group: '' });
	let createGroupOpen: boolean = $state(false);
	let deleteUser: { email: string; name: string; group: string } | null = $state(null);
	let deleteUserOpen = $state(false);

	let canEdit = $derived(
		viewerIsOwner === true || (UserInfo.data?.me.__typename == 'User' && UserInfo.data?.me.isAdmin)
	);

	// let after: string = $state($Groups.variables?.after ?? '');
	// let before: string = $state($Groups.variables?.before ?? '');

	// const changeQuery = (
	// 	params: {
	// 		after?: string;
	// 		before?: string;
	// 	} = {}
	// ) => {
	// 	changeParams({
	// 		before: params.before ?? before,
	// 		after: params.after ?? after
	// 	});
	// };
</script>

<GraphErrors errors={$Groups.errors} />
{#if team}
	<div class="content-wrapper">
		<div>
			{#each team.groups.edges as edge (edge.node.name)}
				{@const memberCount = edge.node.members.pageInfo.totalCount}
				<List title="{edge.node.name}: {memberCount} user{memberCount !== 1 ? 's' : ''}">
					{#snippet menu()}
						{#if canEdit}
							<div class="button">
								<Button
									size="small"
									onclick={() => {
										addMemberProps = {
											open: !addMemberProps.open,
											group: edge.node.name
										};
									}}
									icon={PlusIcon}>Add member</Button
								>
							</div>
						{/if}
						<OrderByMenu
							orderField={GroupMemberOrderField}
							defaultOrderField={GroupMemberOrderField.NAME}
						/>
					{/snippet}
					{#if edge.node.members.edges}
						{#each edge.node.members.edges as memberEdge (memberEdge.node.user.email)}
							<ListItem>
								<div class="item">
									<div>
										<BodyShort size="small">
											{memberEdge.node.user.name}
										</BodyShort>
										<BodyShort size="small">
											<span style="color: var(--ax-text-subtle, --a-text-subtle);"
												>{memberEdge.node.user.email}</span
											>
										</BodyShort>
									</div>

									<div class="role-and-buttons">
										{#if canEdit}
											<div>
												<Button
													title="Delete member"
													size="small"
													variant="tertiary-neutral"
													onclick={() => {
														deleteUser = {
															email: memberEdge.node.user.email,
															name: memberEdge.node.user.name,
															group: edge.node.name
														};
														deleteUserOpen = true;
													}}
												>
													{#snippet icon()}
														<TrashIcon
															style="color:var(--ax-text-danger-icon, --a-icon-danger)!important"
														/>
													{/snippet}
												</Button>
											</div>
										{/if}
									</div>
								</div>
							</ListItem>
						{/each}
					{/if}
				</List>
			{/each}

			{#if canEdit}
				<div class="button">
					<Button
						size="small"
						onclick={() => {
							createGroupOpen = !createGroupOpen;
						}}
						icon={PlusIcon}>Create Group</Button
					>
				</div>
			{/if}
			<!-- <Pagination
				page={$Members.data?.team.members.pageInfo}
				loaders={{
					loadPreviousPage: () => {
						changeQuery({
							before: $Members.data?.team.members.pageInfo.startCursor ?? '',
							after: ''
						});
					},
					loadNextPage: () => {
						changeQuery({
							after: $Members.data?.team.members.pageInfo.endCursor ?? '',
							before: ''
						});
					}
				}}
			/> -->
		</div>
		<!--div>Here be documentation of teams, members and roles</div-->
	</div>
	{#if team}
		<AddMember bind:open={addMemberProps.open} group={addMemberProps.group} on:created={refetch} />

		<CreateGroup bind:open={createGroupOpen} team={team.slug} on:created={refetch} />

		{#if deleteUser && deleteUserOpen}
			{@const group = deleteUser.group}
			{@const userId = deleteUser.email}
			<Confirm
				bind:open={deleteUserOpen}
				confirmText="Delete"
				variant="danger"
				onconfirm={async () => {
					await removeGroupMember.mutate({ input: { groupName: group, userEmail: userId } });
					refetch();
				}}
			>
				{#snippet header()}
					<Heading>Delete Member</Heading>
				{/snippet}
				Are you sure you want to remove <b>{deleteUser.name}</b> from this group?
			</Confirm>
		{/if}
	{/if}
{/if}

<style>
	.button {
		display: flex;
		justify-content: flex-end;
		margin-bottom: var(--ax-space-24, --a-spacing-6);
	}
	.content-wrapper {
		display: grid;
		gap: var(--ax-space-24, --a-spacing-6);
		grid-template-columns: 1fr 300px;
	}

	.item {
		display: grid;
		grid-template-columns: 600px 1fr;
	}
	.role-and-buttons {
		display: flex;
		flex-direction: row;
		justify-content: space-between;
		align-items: center;
		width: 200px;
	}
</style>
