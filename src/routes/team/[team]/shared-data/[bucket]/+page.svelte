<script lang="ts">
	import type { PageProps } from './$types';
	import Pagination from '$lib/ui/Pagination.svelte';
	import type { TeamSharedBucketAccess$result } from '$houdini';
	import UsersTable, { type TeamMemberData } from './UsersTable.svelte';
	import { BodyShort, CopyButton } from '@nais/ds-svelte-community';

	let { params, data }: PageProps = $props();

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

<style>
	.bucket-info {
		display: grid;
		grid-template-columns: 1fr 300px;
		gap: var(--spacing-layout);
		margin-top: calc(-1 * var(--spacing-layout));
		margin-bottom: var(--spacing-layout);
	}
</style>
