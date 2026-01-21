<script lang="ts">
	import type { PageProps } from './$types';
	import Pagination from '$lib/ui/Pagination.svelte';
	import type { TeamSharedBucketAccess$result } from '$houdini';
	import UsersTable, { type TeamMemberData } from './UsersTable.svelte';

	let { data }: PageProps = $props();

	let { TeamSharedBucketAccess } = $derived(data);

	type TeamMemberNode = TeamSharedBucketAccess$result['sharedBucket']['users']['nodes'][0];
	function transformUserData(teamMember: TeamMemberNode): TeamMemberData {
		return {
			...teamMember,
			...{
				user: {
					...teamMember.user,
					...{
						section: {
							name: teamMember.user.section?.name ?? 'Mangler seksjon',
							code: teamMember.user.section?.code ?? '???',
							manager: teamMember.user.section?.manager ?? {
								name: 'Mangler seksjonssjef'
							}
						}
					}
				}
			}
		};
	}
</script>

<div class="container">
	<UsersTable
		teamMembersData={$TeamSharedBucketAccess.data?.sharedBucket.users.nodes.map(
			transformUserData
		) ?? []}
		defaultSelected={data.bucketTableFields}
	/>
</div>

<Pagination
	page={$TeamSharedBucketAccess.data?.sharedBucket.users.pageInfo}
	loaders={{
		loadPreviousPage: () => TeamSharedBucketAccess.loadPreviousPage(),
		loadNextPage: () => TeamSharedBucketAccess.loadNextPage()
	}}
/>
