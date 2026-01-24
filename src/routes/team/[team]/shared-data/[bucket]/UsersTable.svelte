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
		teamMembersData: TeamMemberData[];
		defaultSelected: string[];
	}

	let { teamMembersData, defaultSelected }: Props = $props();
	if (defaultSelected.length == 0) {
		defaultSelected = ['name', 'section', 'team', 'groups'];
	}

	export type TeamMemberData = {
		team: {
			slug: string;
			displayName: string;
		};
		groups: {
			name: string;
		}[];
		user: {
			name: string;
			email: string;
			section: {
				manager: {
					name: string;
					email?: string;
				};
				code: string;
				name: string;
			};
		};
	};

	let selectedFields: string[] = $state(defaultSelected);

	$effect(() => {
		if (!browser) return;
		// group togheter e.g. /users/. Every path start with /
		const path = page.url.pathname.split('/')[1];
		// expiry is 1 year as default in most browsers
		document.cookie = `bucketUsersTableFields${path}=${JSON.stringify(selectedFields)}; expires=Thu, 31 Dec 2099 23:59:59 GMT; SameSite=Lax; Secure; path=/${path}`;
	});

	type SortBy = 'NAME' | 'TEAM' | 'GROUPS';

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

	let teamMembersTable: TeamMemberData[] = $derived(
		getTeamMembersTableDataSorted(teamMembersData, sortState.orderBy, sortState.direction)
	);

	function getTeamMembersTableDataSorted(
		data: TeamMemberData[] | null,
		sortedBy: string,
		sortDirection: string
	): TeamMemberData[] {
		if (!data) {
			return [];
		}

		return [...data].sort((a, b) => {
			let result = 0;
			if (sortedBy === 'NAME') {
				if (a.user.name > b.user.name) result = 1;
				else if (a.user.name < b.user.name) result = -1;
			} else if (sortedBy === 'SECTION') {
				let aS = a.user.section?.code ?? 0;
				let bS = b.user.section?.code ?? 0;
				if (aS > bS) result = 1;
				else if (aS < bS) result = -1;
			} else if (sortedBy === 'TEAM') {
				if (a.team.slug > b.team.slug) result = 1;
				else if (a.team.slug < b.team.slug) result = -1;
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
				<Checkbox value="section">Ansvarlig</Checkbox>
				<Checkbox value="team">Tilgang via</Checkbox>
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
			{#if selectedFields.includes('section')}
				<Th>Ansvarlig</Th>
			{/if}
			{#if selectedFields.includes('team')}
				<Th colspan={2} sortable={true} sortKey="TEAM">Tilgang via</Th>
			{/if}
		</Tr>
	</Thead>
	<Tbody>
		{#each teamMembersTable as teamMember (teamMember.user.email + ':' + teamMember.team.slug)}
			<Tr shadeOnHover={false}>
				<Td>
					<a href={`/user/${teamMember.user.email}/shared-data`}>
						<b>{teamMember.user.name}</b>
					</a>
					<br />
					{teamMember.user.email}
				</Td>
				{#if selectedFields.includes('section')}
					<Td>
						{#if teamMember.user.section.manager.email}
							<a href="/user/{teamMember.user.section.manager.email}"
								>{teamMember.user.section.manager.name}</a
							>
						{:else}
							{teamMember.user.section.manager.name}
						{/if}
						<br />
						{teamMember.user.section.name} ({teamMember.user.section.code})
					</Td>
				{/if}
				{#if selectedFields.includes('team')}
					<Td>
						<a href={`/team/${teamMember.team.slug}/`}>
							<b>{teamMember.team.displayName}</b>
						</a>
						<br />
						{teamMember.team.slug}
					</Td>
					<Td>
						{teamMember.groups.map((g) => g.name.slice(teamMember.team.slug.length + 1)).join(', ')}
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
