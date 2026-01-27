import type { TableSortState } from '@nais/ds-svelte-community';

export interface TeamMemberData {
	id: string;
	user: {
		name: string;
		email: string;
	};
	teamCount: number;
	dataAdminCount: number;
	sectionManager?: {
		name: string;
		email: string;
	};
}

export type SortBy = 'NAME' | 'TEAM_COUNT' | 'DATA_ADMIN_COUNT' | 'SECTION_MANAGER';

export function sortTable(key: SortBy, sortState: TableSortState | null): TableSortState {
	if (!sortState) {
		return {
			orderBy: key,
			direction: 'descending'
		};
	}

	if (sortState.orderBy === key) {
		return {
			...sortState,
			direction: sortState.direction === 'ascending' ? 'descending' : 'ascending'
		};
	}

	return {
		orderBy: key,
		direction: sortState.direction === 'ascending' ? 'descending' : 'ascending'
	};
}

export function getTeamMembersSorted(
	data: TeamMemberData[],
	sortedBy: string,
	sortDirection: string
): TeamMemberData[] {
	if (!data) {
		return [];
	}

	return [...data].sort((a, b) => {
		let result = 0;
		if (sortedBy === 'NAME') {
			result = a.user.name.localeCompare(b.user.name);
		} else if (sortedBy === 'TEAM_COUNT') {
			result = a.teamCount - b.teamCount;
		} else if (sortedBy === 'DATA_ADMIN_COUNT') {
			result = a.dataAdminCount - b.dataAdminCount;
		} else if (sortedBy === 'SECTION_MANAGER') {
			const aName = a.sectionManager?.name ?? '';
			const bName = b.sectionManager?.name ?? '';
			result = aName.localeCompare(bName);
		}

		return sortDirection === 'descending' ? -result : result;
	});
}
