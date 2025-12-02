<script lang="ts">
	import Pagination from '$lib/Pagination.svelte';
	import { Heading, Table, Tbody, Td, Th, Thead, Tr } from '@nais/ds-svelte-community';
	import type { PageProps } from './$houdini';

	let { data }: PageProps = $props();

	let { AdminUsers } = $derived(data);
</script>

<Heading level="2" size="medium" spacing>Brukere</Heading>
{#if $AdminUsers.data}
	<Table size="small">
		<Thead>
			<Tr>
				<Th>Navn</Th>
				<Th>E-post</Th>
				<Th>Ekstern ID</Th>
				<Th>Nais-administrator</Th>
			</Tr>
		</Thead>
		<Tbody>
			{#each $AdminUsers.data.users.nodes || [] as user (user.id)}
				<Tr>
					<Td>{user.name}</Td>
					<Td>{user.email}</Td>
					<Td>{user.externalID}</Td>
					<Td>{user.isAdmin ? 'Ja' : ''}</Td>
				</Tr>
			{:else}
				<Tr>
					<Td colspan={99}>Ingen brukere funnet</Td>
				</Tr>
			{/each}
		</Tbody>
	</Table>

	<Pagination
		page={$AdminUsers.data.users.pageInfo}
		loaders={{
			loadNextPage: () => {
				AdminUsers.loadNextPage();
			},
			loadPreviousPage: () => {
				AdminUsers.loadPreviousPage();
			}
		}}
	/>
{/if}
