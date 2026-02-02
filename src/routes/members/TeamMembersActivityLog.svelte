<script lang="ts">
	import { ActivityLogActivityType, graphql, type ActivityLogFilter } from '$houdini';
	import { Heading, Loader, Tooltip } from '@nais/ds-svelte-community';
	import { RocketIcon } from '@nais/ds-svelte-community/icons';
	import type { Component } from 'svelte';

	import { icons } from '$lib/domain/activity/activity-log-icons';
	import { activityTooltip } from '$lib/domain/activity/activity-log-tooltip';
	import '$lib/domain/activity/activity-log.css';
	import DefaultText from '$lib/domain/activity/shared/texts/DefaultText.svelte';
	import GroupMemberAddedActivityLogEntryText from '$lib/domain/activity/shared/texts/GroupMemberAddedActivityLogEntryText.svelte';

	interface Props {
		teamMembers: Array<{
			user: {
				id: string;
				name: string;
				email: string;
			};
		}>;
		userTeamSlugs: string[];
	}

	let { teamMembers, userTeamSlugs }: Props = $props();

	const filter: ActivityLogFilter = {
		activityTypes: [ActivityLogActivityType.GROUP_MEMBER_ADDED]
	};

	const activityLogQuery = graphql(`
		query TeamMembersActivityLog($filter: ActivityLogFilter) {
			activityLog(first: 15, filter: $filter) {
				edges {
					node {
						id
						actor
						createdAt
						resourceName
						teamSlug
						... on GroupMemberAddedActivityLogEntry {
							__typename
							groupMemberAdded: data {
								userEmail
							}
						}
					}
				}
			}
		}
	`);

	$effect(() => {
		activityLogQuery.fetch({ variables: { filter } });
	});

	let filteredActivities = $derived.by(() => {
		if (!$activityLogQuery.data) {
			return [];
		}

		const memberEmails = new Set(teamMembers.map((m) => m.user.email));
		const teamSlugSet = new Set(userTeamSlugs);

		return (
			$activityLogQuery.data.activityLog.edges
				?.map((edge) => edge.node)
				.filter((entry): entry is typeof entry & { groupMemberAdded: { userEmail: string } } => {
					if (entry.__typename !== 'GroupMemberAddedActivityLogEntry') {
						return false;
					}
					const userEmail = entry.groupMemberAdded?.userEmail;
					if (!userEmail || !memberEmails.has(userEmail)) {
						return false;
					}

					if (!entry.teamSlug || !teamSlugSet.has(entry.teamSlug)) {
						return false;
					}

					if (
						!entry.resourceName ||
						(!entry.resourceName.endsWith('-developers') &&
							!entry.resourceName.endsWith('-data-admins'))
					) {
						return false;
					}

					return true;
				})
				.slice(0, 10) || []
		);
	});

	function textComponent(kind: string): Component<{ data: unknown }> {
		switch (kind) {
			case 'GroupMemberAddedActivityLogEntry':
				return GroupMemberAddedActivityLogEntryText as Component<{ data: unknown }>;
			default:
				return DefaultText as Component<{ data: unknown }>;
		}
	}
</script>

<div class="wrapper">
	<Heading level="2" as="h2" size="small" spacing>Aktivitetslogg</Heading>
	{#if $activityLogQuery.fetching || !$activityLogQuery.data}
		<div style="display: flex; justify-content: center; align-items: center; min-height: 500px;">
			<Loader size="3xlarge" />
		</div>
	{:else}
		{#each filteredActivities as entry, i (entry.id)}
			{@const Icon = icons[entry.__typename] || RocketIcon}
			{@const TextComponent = textComponent(entry.__typename)}
			{@const isLast = i === filteredActivities.length - 1}
			<div class="item" class:last-item={isLast}>
				<div class="activity-icon">
					<Tooltip content={activityTooltip(entry.__typename)}>
						<Icon size="1em" width="1em" height="1em" />
					</Tooltip>
				</div>
				<div class="content">
					<TextComponent data={entry} />
				</div>
			</div>
		{/each}

		{#if !$activityLogQuery.fetching && filteredActivities.length === 0}
			<p class="empty">Ingen nylige aktiviteter funnet.</p>
		{/if}
	{/if}
</div>

<style>
	.wrapper {
		display: flex;
		flex-direction: column;
		gap: var(--ax-space-4);
	}

	.item {
		display: flex;
		position: relative;
		padding-bottom: var(--ax-space-4);
		gap: var(--ax-space-8);
		align-items: flex-start;
	}

	.item:not(.last-item)::before {
		content: '';
		position: absolute;
		left: 16px;
		top: 20px;
		bottom: 0;
		width: 2px;
		background: var(--ax-border-neutral-subtleA);
		z-index: 0;
	}

	.item :global(.activity-icon) {
		background: var(--ax-bg-default);
		border-radius: 50%;
	}

	.content {
		flex: 1;
		min-width: 0;
		max-width: 100%;
		padding-top: var(--ax-space-1);
		overflow-wrap: normal;
		word-break: normal;
	}

	.empty {
		text-align: center;
		color: var(--ax-text-subtle);
		padding: var(--ax-space-8) var(--ax-space-4);
		font-style: italic;
	}
</style>
