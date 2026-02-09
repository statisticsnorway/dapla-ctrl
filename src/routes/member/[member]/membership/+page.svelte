<script lang="ts">
	import Pagination from '$lib/ui/Pagination.svelte';
	import { Heading } from '@nais/ds-svelte-community';
	import type { PageProps } from './$types';
	import {
		graphql,
		type GetUserTeamsForExport$result,
		type UserMemberships$result
	} from '$houdini';
	import DaplaTable from '$lib/ui/DaplaTable.svelte';

	let { data, params }: PageProps = $props();

	let { UserMemberships } = $derived(data);

	type TeamNode = UserMemberships$result['user']['teams']['nodes'][0];

	type TeamItem = TeamNode & { id: string };

	let userTeamsCount = $derived($UserMemberships.data?.user.teams.nodes.length || 0);

	const getAllForExport = graphql(`
		query GetUserTeamsForExport($user: String, $total: Int) {
			user(email: $user) {
				id
				teams(first: $total) {
					nodes {
						groups {
							name
						}
						team {
							slug
							displayName
							isManaged
							section {
								id
								code
								name
								manager {
									name
									email
								}
							}
							members(first: 1) {
								pageInfo {
									totalCount
								}
							}
						}
					}
				}
			}
		}
	`);

	function transformToExportable(
		data: GetUserTeamsForExport$result['user']['teams']['nodes']
	): object[] {
		return data.map((n) => {
			return {
				slug: n.team.slug,
				autonomy: n.team.isManaged ? 'MANAGED' : 'SELF_MANAGED',
				sectionCode: n.team.section.code,
				sectionName: n.team.section.name,
				manager: n.team.section.manager?.email,
				memberCount: n.team.members.pageInfo.totalCount,
				groups: n.groups.map((g) => g.name.substring(n.team.slug.length + 1))
			};
		});
	}
</script>

{#snippet nameCell(item: TeamItem)}
	<a href={`/team/${item.team.slug}/`}>
		<b>{item.team.displayName}</b>
	</a>
	<br />
	{item.team.slug}
{/snippet}
{#snippet autonomyCell(item: TeamItem)}
	{#if item.team.isManaged}
		Managed
	{:else}
		Self-managed
	{/if}
{/snippet}
{#snippet rolesCell(item: TeamItem)}
	{#if item.team.section.manager?.email === params.member}
		<i>Teamansvarlig</i>{#if item.groups.length > 0},
		{/if}
	{/if}
	{item.groups
		?.map((g) => g.name.substring(item.team.slug.length + 1))
		.toSorted()
		.join(', ') ?? ''}
{/snippet}
{#snippet membersCell(item: TeamItem)}
	{item.team.members.pageInfo.totalCount}
{/snippet}
{#snippet managerCell(item: TeamItem)}
	{#if item.team.section.manager}
		<a href="/member/{item.team.section.manager.email}">{item.team.section.manager.name}</a>
	{:else}
		Mangler seksjonsleder
	{/if}
	<br />
	{item.team.section.name} ({item.team.section.code})
{/snippet}

{#if $UserMemberships.data?.user?.teams?.nodes}
	<div class="container">
		<div>
			<div class="section-header">
				<Heading level="2" as="h2" spacing>Medlem av {userTeamsCount} team</Heading>
			</div>
			<DaplaTable
				exportTable={() =>
					getAllForExport
						.fetch({
							variables: {
								user: params.member,
								total: $UserMemberships.data?.user.teams.pageInfo.totalCount
							}
						})
						.then((result) => transformToExportable(result.data?.user.teams.nodes ?? []))}
				data={$UserMemberships.data.user.teams.nodes.map((n) => {
					return { id: n.team.id, ...n };
				})}
				fieldsCookie={{ path: '/member' }}
				selected={data.teamTableFields ?? []}
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
						id: 'ROLES',
						show: 'DEFAULT_YES',
						name: 'Roller',
						cell: rolesCell
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
	</div>
	{#if $UserMemberships.data?.user?.teams?.pageInfo}
		<Pagination
			page={$UserMemberships.data.user.teams.pageInfo}
			loaders={{
				loadPreviousPage: () => UserMemberships.loadPreviousPage(),
				loadNextPage: () => UserMemberships.loadNextPage()
			}}
		/>
	{/if}
{/if}
