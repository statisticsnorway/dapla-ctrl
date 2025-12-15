<script lang="ts">
	import { Table, type TableSortState, Tbody, Td, Th, Thead, Tr } from '@nais/ds-svelte-community';

	interface Props {
		teamsData: TeamsData[];
	}

	let { teamsData }: Props = $props();

	export type User = {
		name: string;
		email: string;
	};

	export type TeamsData = {
		slug: string;
		memberCount: number;
		managers: User[];
		section: {
			code: string;
			name: string;
		};
	};

	type SortBy = 'NAME' | 'MEMBER_COUNT' | 'MANAGER';

	const sortTable = (key: SortBy, sortState: TableSortState) => {
		if (!sortState) {
			sortState = {
				orderBy: key,
				direction: 'descending'
			};
		} else if (sortState.orderBy === key) {
			if (sortState.direction === 'ascending') {
				sortState.direction = 'descending';
			} else {
				sortState.direction = 'ascending';
			}
		} else {
			sortState.orderBy = key;
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

	let teamsTable: TeamsData[] = $derived(
		getTeamsTableDataSorted(teamsData, sortState.orderBy, sortState.direction)
	);

	function getTeamsTableDataSorted(
		data: TeamsData[] | null,
		sortedBy: string,
		sortDirection: string
	): TeamsData[] {
		if (!data) {
			return [];
		}

		return [...data].sort((a, b) => {
			if (sortedBy === 'NAME') {
				if (sortDirection === 'descending') {
					if (a.slug > b.slug) return -1;
					if (a.slug < b.slug) return 1;
					return 0;
				} else {
					if (a.slug > b.slug) return 1;
					if (a.slug < b.slug) return -1;
					return 0;
				}
			} else if (sortedBy === 'MEMBER_COUNT') {
				if (sortDirection === 'descending') {
					if (a.memberCount > b.memberCount) return -1;
					if (a.memberCount < b.memberCount) return 1;
					return 0;
				} else {
					if (a.memberCount > b.memberCount) return 1;
					if (a.memberCount < b.memberCount) return -1;
					return 0;
				}
			} else if (sortedBy === 'MANAGER') {
				if (sortDirection === 'descending') {
					if (a.managers > b.managers) return -1;
					if (a.managers < b.managers) return 1;
					return 0;
				} else {
					if (a.managers > b.managers) return 1;
					if (a.managers < b.managers) return -1;
					return 0;
				}
			}
			return 0;
		});
	}
</script>

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
					<br />
					{team.section.name} ({team.section.code})
				</Td>
				<Td>
					{team.memberCount}
				</Td>
				<Td
					>{#each team.managers as manager, i (manager.email)}
						<a href="/bruker/${manager.email}"> {manager.name}</a>{#if i < team.managers.length - 1}
							<span>,</span>
						{/if}
					{/each}</Td
				>
			</Tr>
		{/each}
	</Tbody>
</Table>

<style>
	span::after {
		/* Avoid 'svelte/no-useless-mustaches' by adding space after comma with css*/
		content: ' ';
	}
</style>
