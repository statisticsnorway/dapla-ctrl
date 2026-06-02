<script module lang="ts">
	type ValueOf<T> = T[keyof T];

	export type OrderField = {
		[s: string]: string;
	};

	export const urlToOrderField = <T extends OrderField>(
		orderField: T,
		defaultValue: ValueOf<T>,
		url: URL
	): ValueOf<T> =>
		(Object.values(orderField).find((field) =>
			url.searchParams.get('sort')?.toUpperCase().startsWith(field.toUpperCase())
		) as ValueOf<T> | undefined) ?? defaultValue;

	export const urlToOrderDirection = (
		url: URL,
		defaultDirection: OrderDirection$options = OrderDirection.ASC
	) =>
		Object.values(OrderDirection).find((dir) =>
			url.searchParams.get('sort')?.toUpperCase().endsWith(dir.toUpperCase())
		) ?? defaultDirection;
</script>

<script lang="ts" generics="Item extends { id: string }">
	import { browser } from '$app/environment';
	import {
		Button,
		Checkbox,
		CheckboxGroup,
		Table,
		Tbody,
		Td,
		Th,
		Thead,
		Tr,
		type TableSortState
	} from '@nais/ds-svelte-community';
	import { ActionMenu, ActionMenuGroup } from '@nais/ds-svelte-community/experimental';
	import { CloudDownIcon, SidebarBothIcon } from '@nais/ds-svelte-community/icons';
	import { page } from '$app/state';
	import type { Snippet } from 'svelte';
	import Papa from 'papaparse';
	import { OrderDirection, type OrderDirection$options } from '$houdini';
	import { changeParams } from '$lib/utils/searchparams';

	type Show = 'ALWAYS' | 'DEFAULT_NO' | 'DEFAULT_YES';

	type Column = {
		id: string;
		show: Show;
		name: string;
		heading?: string | Snippet;
		colspan?: number;
		cell: Snippet<[Item]> | { id: string; align?: 'left' | 'right'; snippet: Snippet<[Item]> }[];
		align?: 'right' | 'left';
		sortKey?: string;
	};

	interface Props {
		data: Item[];
		columns: Column[];
		selected: string[];
		fieldsCookie?: {
			path?: string;
			key?: string;
		};
		exportTable?: () => Promise<object[]>;
	}

	let { data, columns, selected, fieldsCookie, exportTable }: Props = $props();

	// svelte-ignore state_referenced_locally
	let selectedFields: string[] = $state(
		selected.length === 0
			? columns.filter((c) => c.show !== 'DEFAULT_NO').map((c) => c.id)
			: selected
	);

	let selectable = $derived(columns.filter((c) => c.show !== 'ALWAYS'));

	let selectedColumns = $derived.by(() => {
		return columns.filter((c) => selectedFields.includes(c.id));
	});

	const orderByFields = $derived(columns.map((c) => c.sortKey).filter((f) => f !== undefined));

	const orderField = $derived(
		orderByFields.find(
			(f) => page.url.searchParams.get('sort')?.split('-')[0].toUpperCase() === f.toUpperCase()
		) ?? undefined
	);
	const orderDirection = $derived(urlToOrderDirection(page.url));

	let sortState: TableSortState | undefined = $derived(
		orderField !== undefined
			? {
					orderBy: orderField,
					direction: orderDirection === OrderDirection.ASC ? 'ascending' : 'descending'
				}
			: undefined
	);

	$effect(() => {
		if (!browser) return;

		const path = fieldsCookie?.path ?? page.url.pathname.split('/')[1];
		const cookieKey = fieldsCookie?.key ?? `daplaTableFields${path}`;
		const cookiePath = path[0] === '/' ? path : `/${path}`;

		document.cookie = `${cookieKey}=${JSON.stringify(selectedFields)}; expires=Thu, 31 Dec 2099 23:59:59 GMT; SameSite=Lax; Secure; path=${cookiePath}`;
	});

	function toQueryValue(sortState?: TableSortState): string {
		if (!sortState) return '';
		return `${sortState.orderBy}-${sortState.direction === 'ascending' ? 'ASC' : 'DESC'}`;
	}

	async function generateDownload() {
		if (!exportTable) return;
		const csv = Papa.unparse(await exportTable(), { delimiter: ';' });
		const blob = new Blob([csv], { type: `text/csv;charset=utf-8` });
		const link = document.createElement('a');
		const crumbs = page.data.meta.breadcrumbs?.map((b) => b.label).join('_');
		const title = page.data.meta.title;
		const date = new Date().toJSON();
		const filename = `${[crumbs, title, date].filter((e) => e !== undefined).join('_')}.csv`;
		if (link.download !== undefined) {
			const url = URL.createObjectURL(blob);
			link.setAttribute('href', url);
			link.setAttribute('download', filename);
			link.style.visibility = 'hidden';
			document.body.appendChild(link);
			link.click();
			document.body.removeChild(link);
		}
	}
</script>

{#if selectable.length > 0}
	<div class="field-selector">
		{#if exportTable}
			<Button
				variant="tertiary-neutral"
				size="small"
				iconPosition="right"
				icon={CloudDownIcon}
				onclick={generateDownload}
				title="Last ned som CSV"
			></Button>
		{/if}
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
					{#each selectable as column (column.id)}
						<Checkbox value={column.id}>{column.name}</Checkbox>
					{/each}
				</CheckboxGroup>
			</ActionMenuGroup>
		</ActionMenu>
	</div>
{/if}

<Table
	zebraStripes
	sort={sortState}
	onsortchange={(key) => {
		if (!sortState) {
			sortState = {
				orderBy: key,
				direction: 'ascending'
			};
		} else if (sortState.orderBy === key) {
			if (sortState.direction === 'descending') {
				sortState = undefined;
			} else if (sortState.direction === 'ascending') {
				sortState.direction = 'descending';
			} else {
				sortState.direction = 'ascending';
			}
		} else {
			sortState.orderBy = key;
			sortState.direction = 'ascending';
		}

		changeParams({
			sort: toQueryValue(sortState),
			after: '',
			before: ''
		});
	}}
>
	<Thead>
		<Tr>
			{#each selectedColumns as column (column.id)}
				<Th
					sortable={column.sortKey !== undefined}
					sortKey={column.sortKey}
					colspan={(column.colspan ?? typeof column.cell !== 'function') ? column.cell.length : 1}
					align={column.align ?? 'left'}>{column.heading ?? column.name}</Th
				>
			{/each}
		</Tr>
	</Thead>
	<Tbody>
		{#each data as item (item.id)}
			<Tr shadeOnHover={false}>
				{#each selectedColumns as column (column.id)}
					{@const cols =
						typeof column.cell === 'function'
							? [{ id: 'cell', snippet: column.cell }]
							: column.cell}
					{#each cols as col (col.id)}
						<Td align={col.align ?? column.align ?? 'left'}>
							{@render col.snippet(item)}
						</Td>
					{/each}
				{/each}
			</Tr>
		{/each}
	</Tbody>
</Table>

<style>
	.field-selector {
		float: right;
		display: flex;
	}
</style>
