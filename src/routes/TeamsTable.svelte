<script lang="ts">
	import { Table, type TableSortState, Tbody, Td, Th, Thead, Tr } from '@nais/ds-svelte-community';

	interface Props {
		teamsData: TeamsData[];
		rolesHeading?: string;
	}

	let { teamsData, rolesHeading }: Props = $props();

	export type User = {
		name: string;
		email: string;
	};

	export type TeamsData = {
		slug: string;
		purpose: string;
		memberCount: number;
		manager: User;
		section: {
			code: string;
			name: string;
		};
		userGroups?: string[];
	};

	const hasUserGroups = teamsData.some((t) => t.userGroups !== undefined);

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
					if (a.manager > b.manager) return -1;
					if (a.manager < b.manager) return 1;
					return 0;
				} else {
					if (a.manager > b.manager) return 1;
					if (a.manager < b.manager) return -1;
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
			{#if hasUserGroups}
				<Th>{rolesHeading ?? 'Mine roller'}</Th>
			{/if}
			<Th sortable={true} sortKey="MEMBER_COUNT" align="right">Teammedlemmer</Th>
			<Th sortable={true} sortKey="MANAGER">Ansvarlig</Th>
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
					{team.purpose}
				</Td>
				{#if team.userGroups}
					<Td>
						{team.userGroups
							.map((g) => g.substring(team.slug.length + 1))
							.toSorted()
							.join(', ')}
					</Td>
				{/if}
				<Td align="right">
					{team.memberCount}
				</Td>
				<Td>
					{#if team.manager.email !== ''}
						<a href="/user/{team.manager.email}">{team.manager.name}</a>
					{:else}
						{team.manager.name}
					{/if}
					<br />
					{team.section.name} ({team.section.code})
				</Td>
			</Tr>
		{/each}
	</Tbody>
</Table>
