<script lang="ts">
	import {
		ActivityLogActivityType,
		type ActivityLog$input,
		type ActivityLogActivityType$options
	} from '$houdini';
	import ActivityLogItem from '$lib/domain/list-items/ActivityLogListItem.svelte';
	import List from '$lib/ui/List.svelte';
	import Pagination from '$lib/ui/Pagination.svelte';
	import { capitalizeFirstLetter } from '$lib/utils/formatters';
	import { BodyLong, Button, Search } from '@nais/ds-svelte-community';
	import { ActionMenu, ActionMenuCheckboxItem } from '@nais/ds-svelte-community/experimental';
	import { ChevronDownIcon } from '@nais/ds-svelte-community/icons';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let { ActivityLog, teamSlug } = $derived(data);

	const allActivities = Object.values(ActivityLogActivityType).map((type) => type);

	let filtered = $state(allActivities);

	let allActivitiesButtonState: boolean | 'indeterminate' = $derived(
		filtered.length === allActivities.length
			? true
			: filtered.length === 0
				? false
				: 'indeterminate'
	);

	let searchQuery = $state('');

	const groupedActivities: Record<string, ActivityLogActivityType$options[]> = {
		Gruppe: [
			ActivityLogActivityType.GROUP_CREATED,
			ActivityLogActivityType.GROUP_MEMBER_ADDED,
			ActivityLogActivityType.GROUP_MEMBER_REMOVED
		],
		Team: [ActivityLogActivityType.TEAM_CREATED, ActivityLogActivityType.TEAM_UPDATED]
	};

	const activityTypeToNorweigan = (activityType: ActivityLogActivityType$options) => {
		switch (activityType) {
			case 'GROUP_CREATED':
				return 'Gruppe opprettet';
			case 'GROUP_MEMBER_ADDED':
				return 'Gruppemedlem lagt til';
			case 'GROUP_MEMBER_REMOVED':
				return 'Gruppemedlem fjernet';
			case 'TEAM_CREATED':
				return 'Team opprettet';
			case 'TEAM_UPDATED':
				return 'Team oppdatert';
			default:
				return capitalizeFirstLetter(activityType.split('_').join(' ').toLowerCase());
		}
	};

	function filteredGroup(types: string[]) {
		if (!searchQuery) return types;
		return types.filter((type) => type.toLowerCase().includes(searchQuery.toLowerCase()));
	}

	function filterActivities() {
		ActivityLog.fetch({
			variables: {
				team: teamSlug.valueOf(),
				first: 20,
				after: undefined,
				filter: {
					activityTypes: filtered.length === allActivities.length ? [] : filtered.toSorted()
				}
			} as ActivityLog$input
		});
	}
</script>

<div>
	{#if $ActivityLog.data}
		{@const ae = $ActivityLog.data.team.activityLog}
		<div class="wrapper">
			<div>
				<BodyLong spacing
					>Aktivitetsloggen gir et overblikk over hvilke endringer som er gjort på ditt team.</BodyLong
				>
				<List title="{ae.pageInfo.totalCount} hendelser">
					{#snippet menu()}
						<ActionMenu>
							{#snippet trigger(props)}
								<Button
									variant="tertiary-neutral"
									size="small"
									iconPosition="right"
									{...props}
									icon={ChevronDownIcon}
								>
									<span style="font-weight: normal">Filter</span>
								</Button>
							{/snippet}
							<div class="activity-search-wrapper">
								<Search
									class="activity-filter-search"
									placeholder="Søk etter aktivitetstype…"
									label="Søk etter aktivitetstype"
									size="small"
									bind:value={searchQuery}
								/>
							</div>
							<ActionMenuCheckboxItem
								checked={allActivitiesButtonState}
								onchange={(checked) => {
									filtered = checked ? [...allActivities] : [];
									filterActivities();
								}}
							>
								Alle aktiviteter
							</ActionMenuCheckboxItem>
							{#each Object.entries(groupedActivities) as [group, types] (group)}
								{#if filteredGroup(types).length}
									<div class="activity-group-label">{group}</div>
									{#each filteredGroup(types) as type (type)}
										<ActionMenuCheckboxItem
											checked={filtered.includes(type as ActivityLogActivityType$options)}
											onchange={(checked) => {
												const t = type as ActivityLogActivityType$options;
												filtered = checked ? [...filtered, t] : filtered.filter((a) => a !== t);

												filterActivities();
											}}
										>
											{activityTypeToNorweigan(type as ActivityLogActivityType$options)}
										</ActionMenuCheckboxItem>
									{/each}
								{/if}
							{/each}
						</ActionMenu>
					{/snippet}
					{#each ae.nodes || [] as item (item.id)}
						<ActivityLogItem {item} />
					{/each}
				</List>
				{#if $ActivityLog.data.team.activityLog.pageInfo.hasPreviousPage || $ActivityLog.data.team.activityLog.pageInfo.hasNextPage}
					<Pagination
						page={ae.pageInfo}
						loaders={{
							loadNextPage: () => {
								ActivityLog.loadNextPage({ first: 20 });
							},
							loadPreviousPage: () => {
								ActivityLog.loadPreviousPage({
									last: 20
								});
							}
						}}
					/>
				{/if}
			</div>
		</div>
	{/if}
</div>

<style>
	.wrapper {
		display: grid;
		grid-template-columns: 1fr 300px;
		gap: var(--spacing-layout);
	}

	.activity-search-wrapper {
		padding: var(--ax-space-8);
	}

	.activity-group-label {
		padding: var(--ax-space-4) var(--ax-space-8);
		font-weight: 500;
		color: var(--ax-text-neutral-subtle);
		margin-top: var(--ax-space-2);
	}
</style>
