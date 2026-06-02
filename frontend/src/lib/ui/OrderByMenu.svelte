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
		(Object.values(orderField).find((field) => url.searchParams.get('sort')?.startsWith(field)) as
			| ValueOf<T>
			| undefined) ?? defaultValue;

	export const urlToOrderDirection = (
		url: URL,
		defaultDirection: OrderDirection$options = OrderDirection.ASC
	) =>
		Object.values(OrderDirection).find((dir) => url.searchParams.get('sort')?.endsWith(dir)) ??
		defaultDirection;
</script>

<script lang="ts" generics="T extends OrderField">
	import { page } from '$app/state';
	import { OrderDirection, type OrderDirection$options } from '$houdini';
	import { changeParams } from '$lib/utils/searchparams';
	import { Button } from '@nais/ds-svelte-community';
	import {
		ActionMenu,
		ActionMenuDivider,
		ActionMenuRadioGroup,
		ActionMenuRadioItem
	} from '@nais/ds-svelte-community/experimental';
	import { ChevronDownIcon, SortDownIcon, SortUpIcon } from '@nais/ds-svelte-community/icons';

	interface Props {
		orderField: T;
		defaultOrderField: ValueOf<T>;
		defaultOrderDirection?: OrderDirection$options;
		onlyInclude?: ValueOf<T>[];
	}

	const {
		orderField,
		defaultOrderField,
		defaultOrderDirection = OrderDirection.ASC,
		onlyInclude
	}: Props = $props();

	const currentOrderField = $derived(
		Object.values(orderField).find((field) =>
			page.url.searchParams.get('sort')?.startsWith(field)
		) ?? defaultOrderField
	);

	const orderDirection = $derived(
		Object.values(OrderDirection).find((dir) => page.url.searchParams.get('sort')?.endsWith(dir)) ??
			defaultOrderDirection
	);

	export const orderFieldWeights: Record<string, number> = {
		NAME: 0,
		SLUG: 10
	};

	const fieldLabel = (fieldName: string) => {
		switch (fieldName) {
			case 'SLUG':
				return 'Team';
			case 'LAST_MODIFIED_AT':
				return 'Sist endret';
			case 'RESOURCE_NAME':
				return 'Ressursnavn';
			case 'RESOURCE_TYPE':
				return 'Ressurstype';
			case 'NAME':
				return 'Navn';
			case 'EMAIL':
				return 'E-post';
			default:
				return fieldName.charAt(0).toUpperCase() + fieldName.slice(1).toLowerCase();
		}
	};
</script>

<ActionMenu>
	{#snippet trigger(props)}
		<Button
			variant="tertiary-neutral"
			size="small"
			iconPosition="right"
			icon={ChevronDownIcon}
			{...props}
		>
			<div style="display: flex; align-items: center; gap: var(--ax-space-8);">
				{#if orderDirection === OrderDirection.ASC}
					<SortUpIcon />
				{:else}
					<SortDownIcon />
				{/if}
				{fieldLabel(currentOrderField)}
			</div>
		</Button>
	{/snippet}
	{#key orderField}
		<!-- prettier-ignore -->
		<ActionMenuRadioGroup value={currentOrderField} label="Sorter etter">
			{#each Object.values(orderField)
				.filter((field) => !onlyInclude || onlyInclude.includes(field as ValueOf<T>))
				.sort((a, b) => {
					const aWeight = orderFieldWeights[a as string] ?? 9999;
					const bWeight = orderFieldWeights[b as string] ?? 9999;
					return aWeight - bWeight;
				}) as field (field)}
				<ActionMenuRadioItem
					value={field}
					onselect={(value) =>
						changeParams(
							{ sort: `${value}-${orderDirection}`, after: '', before: '' },
							{ noScroll: true }
						)}
				>
					{fieldLabel(field)}
				</ActionMenuRadioItem>
			{/each}
		</ActionMenuRadioGroup>
	{/key}
	<ActionMenuDivider />
	{#key orderDirection}
		<ActionMenuRadioGroup value={orderDirection} label="Sorteringsrekkefølge">
			{#each Object.values(OrderDirection) as direction (direction)}
				<ActionMenuRadioItem
					value={direction}
					onselect={(value) =>
						changeParams(
							{ sort: `${currentOrderField}-${value}`, after: '', before: '' },
							{ noScroll: true }
						)}
				>
					{#if direction === OrderDirection.ASC}
						<SortUpIcon /> Stigende
					{:else}
						<SortDownIcon /> Synkende
					{/if}
				</ActionMenuRadioItem>
			{/each}
		</ActionMenuRadioGroup>
	{/key}
</ActionMenu>
