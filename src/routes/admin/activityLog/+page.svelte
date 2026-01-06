<script lang="ts">
	import GraphErrors from '$lib/ui/GraphErrors.svelte';
	import Pagination from '$lib/ui/Pagination.svelte';
	import Time from '$lib/ui/Time.svelte';
	import { Heading, Table, Tbody, Td, Th, Thead, Tr } from '@nais/ds-svelte-community';
	import type { PageProps } from './$houdini';

	let { data }: PageProps = $props();
	let { ActivityLogs } = $derived(data);
</script>

<Heading level="2" as="h2" size="medium" spacing>Aktivitetslogg</Heading>
<GraphErrors errors={$ActivityLogs.errors} />
{#if $ActivityLogs.data}
	<Table size="small">
		<Thead>
			<Tr>
				<Th>E-post</Th>
				<Th>Beskrivelse</Th>
				<Th>Ressursnavn</Th>
				<Th>Ressurstype</Th>
				<Th>Team</Th>
				<Th>Tidspunkt</Th>
			</Tr>
		</Thead>
		<Tbody>
			{#each $ActivityLogs.data.activityLog.nodes || [] as entry (entry.id)}
				<Tr>
					<Td>
						{#if entry.actor !== ''}
							<a href="/user/{entry.actor}/">{entry.actor}</a>
						{:else}
							-
						{/if}
					</Td>
					<Td>
						{entry.message}
					</Td>
					<Td>
						{entry.resourceName}
					</Td>
					<Td>
						{entry.resourceType}
					</Td>
					<Td>
						{#if entry.teamSlug !== ''}
							<a href="/team/{entry.teamSlug}/">{entry.teamSlug}</a>
						{:else}
							-
						{/if}
					</Td>
					<Td><Time time={entry.createdAt} distance={true} /></Td>
				</Tr>
			{:else}
				<Tr>
					<Td colspan={99}>Fant ingen aktivitetslogger</Td>
				</Tr>
			{/each}
		</Tbody>
	</Table>

	<Pagination
		page={$ActivityLogs.data.activityLog.pageInfo}
		loaders={{
			loadNextPage: () => {
				ActivityLogs.loadNextPage();
			},
			loadPreviousPage: () => {
				ActivityLogs.loadPreviousPage();
			}
		}}
	/>
{/if}
