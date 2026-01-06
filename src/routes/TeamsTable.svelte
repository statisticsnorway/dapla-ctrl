<script lang="ts">
	import { browser } from '$app/environment';
	import {
		Button,
		Checkbox,
		CheckboxGroup,
		Table,
		type TableSortState,
		Tbody,
		Td,
		Th,
		Thead,
		Tr
	} from '@nais/ds-svelte-community';
	import { ActionMenu, ActionMenuGroup } from '@nais/ds-svelte-community/experimental';
	import { SidebarBothIcon } from '@nais/ds-svelte-community/icons';
	import { page } from '$app/state';

	interface Props {
		teamsData: TeamsData[];
		rolesHeading?: string;
		defaultSelected: string[];
	}

	let { teamsData, rolesHeading, defaultSelected }: Props = $props();
	if (defaultSelected.length == 0) {
		defaultSelected = ['groups', 'memberCount', 'manager'];
	}

	export type User = {
		name: string;
		email: string;
	};

	export type TeamsData = {
		slug: string;
		displayName: string;
		purpose: string;
		memberCount: number;
		manager: User;
		section: {
			code: string;
			name: string;
		};
		userGroups?: string[];
	};

	let selectedFields: string[] = $state(defaultSelected);

	$effect(() => {
		if (!browser) return;
		// group togheter e.g. /users/. Every path start with /
		const path = page.url.pathname.split('/')[1];
		// expiry is 1 year as default in most browsers
		document.cookie = `teamTableFields${path}=${JSON.stringify(selectedFields)}; expires=Thu, 31 Dec 2099 23:59:59 GMT; SameSite=Lax; Secure; path=/${path}`;
	});

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
		orderBy: 'NONE',
		direction: 'ascending'
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
			let result = 0;
			if (sortedBy === 'NAME') {
				if (a.slug > b.slug) result = 1;
				else if (a.slug < b.slug) result = -1;
			} else if (sortedBy === 'GROUPS') {
				const aLen = a.userGroups?.length ?? 0;
				const bLen = a.userGroups?.length ?? 0;
				if (aLen > bLen) result = 1;
				else if (aLen < bLen) result = -1;
			} else if (sortedBy === 'MEMBER_COUNT') {
				if (a.memberCount > b.memberCount) result = 1;
				else if (a.memberCount < b.memberCount) result = -1;
			} else if (sortedBy === 'MANAGER') {
				if (a.manager > b.manager) result = 1;
				else if (a.manager < b.manager) result = -1;
			}

			if (sortDirection === 'descending') result *= -1;
			return result;
		});
	}
</script>

<div class="field-selector">
	<ActionMenu align="end">
		{#snippet trigger(props)}
			<Button
				variant="tertiary-neutral"
				size="small"
				iconPosition="right"
				icon={SidebarBothIcon}
				{...props}
			></Button>
		{/snippet}
		<ActionMenuGroup label="Felter">
			<CheckboxGroup legend="" bind:value={selectedFields}>
				{#if hasUserGroups}
					<Checkbox value="groups">{rolesHeading ?? 'Mine roller'}</Checkbox>
				{/if}
				<Checkbox value="memberCount">Teammedlemmer</Checkbox>
				<Checkbox value="manager">Ansvarlig</Checkbox>
			</CheckboxGroup>
		</ActionMenuGroup>
	</ActionMenu>
</div>
<Table
	sort={sortState}
	onsortchange={(key) => {
		sortState = sortTable(key as SortBy, sortState);
	}}
	zebraStripes
>
	<Thead>
		<Tr>
			<Th sortable={true} sortKey="NAME">Navn</Th>
			{#if hasUserGroups && selectedFields.includes('groups')}
				<Th sortable={true} sortKey="GROUPS">{rolesHeading ?? 'Mine roller'}</Th>
			{/if}
			{#if selectedFields.includes('memberCount')}
				<Th sortable={true} sortKey="MEMBER_COUNT" align="right">Teammedlemmer</Th>
			{/if}
			{#if selectedFields.includes('manager')}
				<Th sortable={true} sortKey="MANAGER">Ansvarlig</Th>
			{/if}
		</Tr>
	</Thead>
	<Tbody>
		{#each teamsTable as team (team.slug)}
			<Tr shadeOnHover={false}>
				<Td>
					<a href={`/team/${team.slug}/`}>
						<b>{team.displayName}</b>
					</a>
					<br />
					{team.slug}
				</Td>
				{#if team.userGroups && selectedFields.includes('groups')}
					<Td>
						{team.userGroups
							.map((g) => g.substring(team.slug.length + 1))
							.toSorted()
							.join(', ')}
					</Td>
				{/if}
				{#if selectedFields.includes('memberCount')}
					<Td align="right">
						{team.memberCount}
					</Td>
				{/if}
				{#if selectedFields.includes('manager')}
					<Td>
						{#if team.manager.email !== ''}
							<a href="/user/{team.manager.email}">{team.manager.name}</a>
						{:else}
							{team.manager.name}
						{/if}
						<br />
						{team.section.name} ({team.section.code})
					</Td>
				{/if}
			</Tr>
		{/each}
	</Tbody>
</Table>

<style>
	.field-selector {
		float: right;
	}
</style>
