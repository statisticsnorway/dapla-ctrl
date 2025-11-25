<script lang="ts">
	import { Heading, Table, type TableSortState, Tbody, Td, Th, Thead, Tr } from '@nais/ds-svelte-community';

	interface Props {
		teamsData: TeamsData[];
	}


	let { teamsData }: Props = $props();

	export type TeamsData = {
		slug: string;
		memberCount: number;
		managers: string[];
	};


	type SortBy = 'NAME' | 'MEMBER_COUNT' | 'MANAGER';

	const sortTable = (key: SortBy, sortState: TableSortState) => {
		if (!sortState) {
			sortState = {
				orderBy: key,
				direction: 'descending'
			};
		} else {
			if (sortState.direction === 'ascending') {
				sortState.direction = 'descending';
			} else {
				sortState.direction = 'ascending';
			}
		}

		return sortState;
	};

	let sortState: TableSortState = $state({
		orderBy: 'NAME',
		direction: 'descending'
	});

	let teamsTable: TeamsData[] = $state([]);

	$effect(() => {
		teamsTable = getTeamsTableDataSorted(
			teamsData,
			sortState.orderBy,
			sortState.direction
		);
	});
	//
	// function getTeamsTableDataSorted (
	// 	data: TeamsData | null,
	// 	sortedBy: string,
	// 	sortDirection: string
	// ): TeamsData[] {
	// 	const teamMap = new Map<string, TeamsData>();
	//
	// 	if (!data) {
	// 		return [];
	// 	}
	//
	//
	//
	// 	return Array.from(teamMap.values()).sort((a, b) => {
	// 		if (sortedBy === 'NAME') {
	// 			if (sortDirection === 'descending') {
	// 				if (a.teamSlug > b.teamSlug) return -1;
	// 				if (a.teamSlug < b.teamSlug) return 1;
	// 				return 0;
	// 			} else {
	// 				if (a.teamSlug > b.teamSlug) return 1;
	// 				if (a.teamSlug < b.teamSlug) return -1;
	// 				return 0;
	// 			}
	// 		} else if (sortedBy === 'CPU') {
	// 			if (sortDirection === 'descending') {
	// 				if (a.unusedCpu > b.unusedCpu) return -1;
	// 				if (a.unusedCpu < b.unusedCpu) return 1;
	// 				return 0;
	// 			} else {
	// 				if (a.unusedCpu > b.unusedCpu) return 1;
	// 				if (a.unusedCpu < b.unusedCpu) return -1;
	// 				return 0;
	// 			}
	// 		} else if (sortedBy === 'MEMORY') {
	// 			if (sortDirection === 'descending') {
	// 				if (a.unusedMem > b.unusedMem) return -1;
	// 				if (a.unusedMem < b.unusedMem) return 1;
	// 				return 0;
	// 			} else {
	// 				if (a.unusedMem > b.unusedMem) return 1;
	// 				if (a.unusedMem < b.unusedMem) return -1;
	// 				return 0;
	// 			}
	// 		}
	// 		return 0;
	// 	});
	// }


</script>

<div class="container">
	<div>
		<Heading level="2" spacing>Mine Team</Heading>

		<Table
			size="small"
			sort={sortState}
			onsortchange={(key) => {
					sortState = sortTable(key as SortBy, sortState);
				}}
		>
			<Thead>
			<Tr>
				<Th sortable={true} sortKey="NAME">Navn</Th>
				<Th sortable={true} sortKey="MEMBER_COUNT">Teammedlemmer</Th>
				<Th sortable={true} sortKey="MANAGER">Managers</Th>
			</Tr>
			</Thead>
			<Tbody>
			{#each teamsTable as team (team.slug)}
				<Tr>
					<Td>
						<a href={`/team/${team.slug}/`}>
							{team.slug}
						</a>
					</Td>
					<Td>
						{team.memberCount}
					</Td>
					<Td>{@html team.managers.map(user => { return `<a href="/user${user.email}">${user.name}</a>`}).join(", ")}</Td>
				</Tr>
			{/each}
			</Tbody>
		</Table>
	</div>
</div>

<style>
    .container {
        margin-top: var(--spacing-layout);
        display: flex;
        flex-direction: column;
        gap: var(--spacing-layout);
    }
</style>
