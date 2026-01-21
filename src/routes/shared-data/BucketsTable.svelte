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
	import { capitalizeFirstLetter } from '$lib/utils/formatters';

	interface Props {
		bucketsData: BucketData[];
		defaultSelected: string[];
	}

	let { bucketsData, defaultSelected }: Props = $props();
	if (defaultSelected.length == 0) {
		defaultSelected = ['name', 'team', 'section', 'type', 'env', 'teams', 'users'];
	}

	export type BucketData = {
		team: {
			slug: string;
			displayName: string;
		};
		name: string;
		shortName: string;
		type: string;
		env: string;
		teamCount: number;
		userCount: number;
		sectionCode: string;
	};

	let selectedFields: string[] = $state(defaultSelected);

	$effect(() => {
		if (!browser) return;
		// group togheter e.g. /users/. Every path start with /
		const path = page.url.pathname.split('/')[1];
		// expiry is 1 year as default in most browsers
		document.cookie = `bucketTableFields${path}=${JSON.stringify(selectedFields)}; expires=Thu, 31 Dec 2099 23:59:59 GMT; SameSite=Lax; Secure; path=/${path}`;
	});

	type SortBy = 'NAME' | 'TEAM' | 'SECTION' | 'TYPE' | 'ENV' | 'TEAMS' | 'USERS';

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

	let bucketsTable: BucketData[] = $derived(
		getBucketsTableDataSorted(bucketsData, sortState.orderBy, sortState.direction)
	);

	function getBucketsTableDataSorted(
		data: BucketData[] | null,
		sortedBy: string,
		sortDirection: string
	): BucketData[] {
		if (!data) {
			return [];
		}

		return [...data].sort((a, b) => {
			let result = 0;
			if (sortedBy === 'NAME') {
				if (a.name > b.name) result = 1;
				else if (a.name < b.name) result = -1;
			} else if (sortedBy === 'TEAMS') {
				if (a.teamCount > b.teamCount) result = 1;
				else if (a.teamCount < b.teamCount) result = -1;
			} else if (sortedBy === 'USERS') {
				if (a.userCount > b.userCount) result = 1;
				else if (a.userCount < b.userCount) result = -1;
			} else if (sortedBy === 'TYPE') {
				if (a.type > b.type) result = 1;
				else if (a.type < b.type) result = -1;
			} else if (sortedBy === 'ENV') {
				if (a.env > b.env) result = 1;
				else if (a.env < b.env) result = -1;
			} else if (sortedBy === 'TEAM') {
				if (a.env > b.env) result = 1;
				else if (a.env < b.env) result = -1;
			} else if (sortedBy === 'SECTION') {
				if (a.sectionCode > b.sectionCode) result = 1;
				else if (a.sectionCode < b.sectionCode) result = -1;
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
				<Checkbox value="type">Type</Checkbox>
				<Checkbox value="team">Team</Checkbox>
				<Checkbox value="section">Seksjon</Checkbox>
				<Checkbox value="env">Miljø</Checkbox>
				<Checkbox value="teams">Antall team</Checkbox>
				<Checkbox value="users">Antall personer</Checkbox>
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
			{#if selectedFields.includes('type')}
				<Th sortable={true} sortKey="TYPE">Type</Th>
			{/if}
			{#if selectedFields.includes('team')}
				<Th sortable={true} sortKey="TEAM">Team</Th>
			{/if}
			{#if selectedFields.includes('section')}
				<Th sortable={true} sortKey="SECTION" align="right">Seksjon</Th>
			{/if}
			{#if selectedFields.includes('env')}
				<Th sortable={true} sortKey="ENV">Miljø</Th>
			{/if}
			{#if selectedFields.includes('teams')}
				<Th sortable={true} sortKey="TEAMS" align="right">Antall team</Th>
			{/if}
			{#if selectedFields.includes('users')}
				<Th sortable={true} sortKey="USERS" align="right">Antall personer</Th>
			{/if}
		</Tr>
	</Thead>
	<Tbody>
		{#each bucketsTable as bucket (bucket.name)}
			<Tr shadeOnHover={false}>
				<Td>
					<a href={`/team/${bucket.team.slug}/shared-data/${bucket.name}`}>
						<b>{bucket.shortName}</b>
					</a>
					<br />
					{bucket.name}
				</Td>
				{#if selectedFields.includes('type')}
					<Td>
						{capitalizeFirstLetter(bucket.type)}
					</Td>
				{/if}
				{#if selectedFields.includes('team')}
					<Td>
						<a href={`/team/${bucket.team.slug}/`}>
							<b>{bucket.team.displayName}</b>
						</a>
						<br />
						{bucket.team.slug}
					</Td>
				{/if}
				{#if selectedFields.includes('section')}
					<Td align="right">
						{bucket.sectionCode}
					</Td>
				{/if}
				{#if selectedFields.includes('env')}
					<Td>
						{bucket.env}
					</Td>
				{/if}
				{#if selectedFields.includes('teams')}
					<Td align="right">
						{bucket.teamCount}
					</Td>
				{/if}
				{#if selectedFields.includes('users')}
					<Td align="right">
						{bucket.userCount}
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
