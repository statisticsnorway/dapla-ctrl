<script lang="ts">
	import { ActivityLogActivityType, graphql, type ActivityLogFilter } from '$houdini';
	import { Heading, Loader, Tooltip } from '@nais/ds-svelte-community';
	import { RocketIcon } from '@nais/ds-svelte-community/icons';
	import type { Component } from 'svelte';

	import { icons } from '../activity-log-icons';
	import { activityTooltip } from '../activity-log-tooltip';
	import '../activity-log.css';
	import DefaultText from '../shared/texts/DefaultText.svelte';
	import GroupCreatedActivityLogEntryText from '../shared/texts/GroupCreatedActivityLogEntryText.svelte';
	import GroupMemberAddedActivityLogEntryText from '../shared/texts/GroupMemberAddedActivityLogEntryText.svelte';
	import GroupMemberRemovedActivityLogEntryText from '../shared/texts/GroupMemberRemovedActivityLogEntryText.svelte';
	import TeamCreatedActivityLogEntryText from '../shared/texts/TeamCreatedActivityLogEntryText.svelte';
	import TeamUpdatedActivityLogEntryText from '../shared/texts/TeamUpdatedActivityLogEntryText.svelte';
	import TeamRoleAssignedActivityLogEntryText from '../shared/texts/TeamRoleAssignedActivityLogEntryText.svelte';
	import TeamRoleRevokedActivityLogEntryText from '../shared/texts/TeamRoleRevokedActivityLogEntryText.svelte';

	interface Props {
		teamSlug: string;
	}
	let { teamSlug }: Props = $props();

	const filter: ActivityLogFilter = {
		activityTypes: [
			ActivityLogActivityType.GROUP_CREATED,
			ActivityLogActivityType.GROUP_MEMBER_ADDED,
			ActivityLogActivityType.GROUP_MEMBER_REMOVED,
			ActivityLogActivityType.RECONCILER_CONFIGURED,
			ActivityLogActivityType.RECONCILER_DISABLED,
			ActivityLogActivityType.RECONCILER_ENABLED,
			ActivityLogActivityType.TEAM_CREATED,
			ActivityLogActivityType.TEAM_UPDATED,
			ActivityLogActivityType.TEAM_ROLE_ASSIGNED,
			ActivityLogActivityType.TEAM_ROLE_REVOKED
		]
	};

	const activityLogQuery = graphql(`
		query TeamOverviewActivityLog($teamSlug: Slug!, $filter: ActivityLogFilter) {
			team(slug: $teamSlug) {
				activityLog(first: 10, filter: $filter) {
					edges {
						node {
							id
							actor
							message
							createdAt
							resourceName
							resourceType
							teamSlug
							... on TeamCreatedActivityLogEntry {
								__typename
							}
							... on GroupCreatedActivityLogEntry {
								__typename
							}
							... on GroupMemberAddedActivityLogEntry {
								__typename

								groupMemberAdded: data {
									userEmail
								}
							}
							... on GroupMemberRemovedActivityLogEntry {
								__typename

								groupMemberRemoved: data {
									userEmail
								}
							}
							... on TeamRoleAssignedActivityLogEntry {
								__typename
								roleAssigned: data {
									role
									user {
										id
										name
										email
									}
								}
							}
							... on TeamRoleRevokedActivityLogEntry {
								__typename
								roleRevoked: data {
									role
									user {
										id
										name
										email
									}
								}
							}
							... on TeamUpdatedActivityLogEntry {
								__typename
								teamUpdated: data {
									updatedFields {
										field
										oldValue
										newValue
									}
								}
							}
							... on ServiceAccountCreatedActivityLogEntry {
								__typename
								resourceName
							}
							... on ServiceAccountTokenCreatedActivityLogEntry {
								__typename
								resourceName
							}
						}
					}
				}
			}
		}
	`);

	$effect(() => {
		activityLogQuery.fetch({ variables: { teamSlug, filter } });
	});

	type Kind = string;

	function textComponent(kind: Kind): Component<{ data: unknown }> {
		switch (kind) {
			case 'TeamCreatedActivityLogEntry':
				return TeamCreatedActivityLogEntryText as Component<{ data: unknown }>;
			case 'TeamUpdatedActivityLogEntry':
				return TeamUpdatedActivityLogEntryText as Component<{ data: unknown }>;
			case 'GroupCreatedActivityLogEntry':
				return GroupCreatedActivityLogEntryText as Component<{ data: unknown }>;
			case 'GroupMemberAddedActivityLogEntry':
				return GroupMemberAddedActivityLogEntryText as Component<{ data: unknown }>;
			case 'GroupMemberRemovedActivityLogEntry':
				return GroupMemberRemovedActivityLogEntryText as Component<{ data: unknown }>;
			case 'TeamRoleAssignedActivityLogEntry':
				return TeamRoleAssignedActivityLogEntryText as Component<{ data: unknown }>;
			case 'TeamRoleRevokedActivityLogEntry':
				return TeamRoleRevokedActivityLogEntryText as Component<{ data: unknown }>;
			default:
				return DefaultText as Component<{ data: unknown }>;
		}
	}
</script>

<div class="wrapper">
	<Heading><a href="/team/{teamSlug}/activity-log">Aktivitetslogg</a></Heading>
	{#if $activityLogQuery.fetching || !$activityLogQuery.data}
		<div style="display: flex; justify-content: center; align-items: center; min-height: 500px;">
			<Loader size="3xlarge" />
		</div>
	{:else}
		{#each $activityLogQuery.data?.team?.activityLog.edges || [] as { node: entry }, i (entry.id)}
			{@const Icon = icons[entry.__typename] || RocketIcon}
			{@const TextComponent = textComponent(entry.__typename)}
			{@const isLast = i === ($activityLogQuery.data?.team?.activityLog.edges?.length ?? 0) - 1}
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

		{#if !$activityLogQuery.fetching && ($activityLogQuery.data?.team?.activityLog.edges || []).length === 0}
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
		gap: var(--ax-space-12);
		align-items: flex-start;
	}

	/* vertical timeline line */
	.item:not(.last-item)::before {
		content: '';
		position: absolute;
		left: 16px; /* centers under 32px icon */
		top: 20px;
		bottom: 0;
		width: 2px;
		background: var(--ax-border-neutral-subtleA);
		z-index: 0;
	}

	/* Add background to icon to block the line */
	.item :global(.activity-icon) {
		background: var(--ax-bg-default);
		border-radius: 50%;
	}

	.content {
		flex: 1;
		min-width: 0;
		padding-top: var(--ax-space-1);
	}

	.empty {
		text-align: center;
		color: var(--ax-text-subtle);
		padding: var(--ax-space-8) var(--ax-space-4);
		font-style: italic;
	}
</style>
