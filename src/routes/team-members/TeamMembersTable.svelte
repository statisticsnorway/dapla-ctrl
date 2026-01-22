<script lang="ts">
	import Pagination from '$lib/ui/Pagination.svelte';
	import { Table, type TableSortState, Tbody, Td, Th, Thead, Tr } from '@nais/ds-svelte-community';
	import {
		getTeamMembersSorted,
		sortTable,
		type SortBy,
		type TeamMemberData
	} from './teamMembersUtils';

	interface Props {
		teamMembers: TeamMemberData[];
		pageInfo?: {
			readonly hasNextPage: boolean;
			readonly hasPreviousPage: boolean;
			readonly pageStart: number;
			readonly pageEnd: number;
			readonly totalCount: number;
		};
		loaders?: {
			loadPreviousPage: () => void;
			loadNextPage: () => void;
		};
	}

	let { teamMembers, pageInfo, loaders }: Props = $props();

	let sortState: TableSortState = $state({
		orderBy: 'NONE',
		direction: 'ascending'
	});

	let sortedTeamMembers = $derived(
		getTeamMembersSorted(teamMembers, sortState.orderBy, sortState.direction)
	);
</script>

<Table
	zebraStripes
	sort={sortState}
	onsortchange={(key) => {
		sortState = sortTable(key as SortBy, sortState);
	}}
>
	<Thead>
		<Tr>
			<Th sortable={true} sortKey="NAME">Navn</Th>
			<Th sortable={true} sortKey="TEAM_COUNT" align="right">Team</Th>
			<Th sortable={true} sortKey="DATA_ADMIN_COUNT" align="right">Data-admin roller</Th>
			<Th sortable={true} sortKey="SECTION_MANAGER">Seksjonsleder</Th>
		</Tr>
	</Thead>
	<Tbody>
		{#each sortedTeamMembers as member (member.user.email)}
			<Tr shadeOnHover={false}>
				<Td>
					<a href="/user/{member.user.email}">{member.user.name}</a>
				</Td>
				<Td align="right">{member.teamCount}</Td>
				<Td align="right">{member.dataAdminCount}</Td>
				<Td>
					{#if member.sectionManager}
						{#if member.sectionManager.email}
							<a href="/user/{member.sectionManager.email}">{member.sectionManager.name}</a>
						{:else}
							{member.sectionManager.name}
						{/if}
					{:else}
						<span style="color: var(--ax-text-subtle); font-style: italic;"
							>Mangler seksjonsleder</span
						>
					{/if}
				</Td>
			</Tr>
		{:else}
			<Tr>
				<Td colspan={4}>Fant ingen teammedlemmer</Td>
			</Tr>
		{/each}
	</Tbody>
</Table>

{#if pageInfo && loaders}
	<Pagination page={pageInfo} {loaders} />
{/if}
