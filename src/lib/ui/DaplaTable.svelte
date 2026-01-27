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
		Tr
	} from '@nais/ds-svelte-community';
	import { ActionMenu, ActionMenuGroup } from '@nais/ds-svelte-community/experimental';
	import { SidebarBothIcon } from '@nais/ds-svelte-community/icons';
	import { page } from '$app/state';
	import type { Snippet } from 'svelte';

	type Show = 'ALWAYS' | 'DEFAULT_NO' | 'DEFAULT_YES';

	type Column = {
		id: string;
		show: Show;
		name: string;
		heading?: string | Snippet;
		colspan?: number;
		cell: Snippet<[Item]> | { id: string; align?: 'left' | 'right'; snippet: Snippet<[Item]> }[];
		align?: 'right' | 'left';
	};

	interface Props {
		data: Item[];
		columns: Column[];
		selected: string[];
		fieldsCookie?: {
			path?: string;
			key?: string;
		};
	}

	let { data, columns, selected, fieldsCookie }: Props = $props();

	if (selected.length === 0) {
		selected = columns.filter((c) => c.show !== 'DEFAULT_NO').map((c) => c.id);
	}
	let selectable = columns.filter((c) => c.show !== 'ALWAYS');

	let selectedFields: string[] = $state(selected);
	let selectedColumns = $derived.by(() => {
		return columns.filter((c) => selectedFields.includes(c.id));
	});

	$effect(() => {
		if (!browser) return;

		const path = fieldsCookie?.path ?? page.url.pathname.split('/')[1];
		const cookieKey = fieldsCookie?.key ?? `daplaTableFields${path}`;
		const cookiePath = path[0] === '/' ? path : `/${path}`;

		document.cookie = `${cookieKey}=${JSON.stringify(selectedFields)}; expires=Thu, 31 Dec 2099 23:59:59 GMT; SameSite=Lax; Secure; path=${cookiePath}`;
	});
</script>

{#if selectable.length > 0}
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
					{#each selectable as column (column.id)}
						<Checkbox value={column.id}>{column.name}</Checkbox>
					{/each}
				</CheckboxGroup>
			</ActionMenuGroup>
		</ActionMenu>
	</div>
{/if}

<Table zebraStripes>
	<Thead>
		<Tr>
			{#each selectedColumns as column (column.id)}
				<Th
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
	}
</style>
