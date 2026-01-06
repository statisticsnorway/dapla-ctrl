<script lang="ts">
	import GraphErrors from '$lib/ui/GraphErrors.svelte';
	import Pagination from '$lib/ui/Pagination.svelte';
	import Time from '$lib/ui/Time.svelte';
	import { Heading, Table, Tbody, Td, Th, Thead, Tr } from '@nais/ds-svelte-community';
	import type { PageProps } from './$houdini';

	let { data }: PageProps = $props();
	let { UserSyncLogs } = $derived(data);
</script>

<Heading level="2" as="h2" size="medium" spacing>Brukersynkroniseringslogg</Heading>
<GraphErrors errors={$UserSyncLogs.errors} />
{#if $UserSyncLogs.data}
	<Table size="small">
		<Thead>
			<Tr>
				<Th>Handling</Th>
				<Th>Navn</Th>
				<Th>E-post</Th>
				<Th>Tidsstempel</Th>
			</Tr>
		</Thead>
		<Tbody>
			{#each $UserSyncLogs.data.userSyncLog.nodes || [] as entry (entry.id)}
				<Tr>
					<Td>
						{#if entry.__typename === 'RoleAssignedUserSyncLogEntry'}
							Tildelt rolle <em>{entry.roleName}</em>
						{:else if entry.__typename === 'RoleRevokedUserSyncLogEntry'}
							Tilbaketrukket rolle <em>{entry.roleName}</em>
						{:else}
							{entry.message}
						{/if}
					</Td>
					<Td>
						{entry.userName}
						{#if entry.__typename === 'UserUpdatedUserSyncLogEntry' && entry.oldUserName !== entry.userName}
							<span class="old-value">{entry.oldUserName}</span>
						{/if}
					</Td>
					<Td>
						{entry.userEmail}
						{#if entry.__typename === 'UserUpdatedUserSyncLogEntry' && entry.oldUserEmail !== entry.userEmail}
							<span class="old-value">{entry.oldUserEmail}</span>
						{/if}
					</Td>
					<Td><Time time={entry.createdAt} distance={true} /></Td>
				</Tr>
			{:else}
				<Tr>
					<Td colspan={99}>Fant ingen brukersynkroniseringslogger</Td>
				</Tr>
			{/each}
		</Tbody>
	</Table>

	<Pagination
		page={$UserSyncLogs.data.userSyncLog.pageInfo}
		loaders={{
			loadNextPage: () => {
				UserSyncLogs.loadNextPage();
			},
			loadPreviousPage: () => {
				UserSyncLogs.loadPreviousPage();
			}
		}}
	/>
{/if}

<style>
	.old-value {
		text-decoration: line-through;
	}
</style>
