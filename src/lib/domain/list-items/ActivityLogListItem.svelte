<script lang="ts">
	import { fragment, graphql, type ActivityLogEntryFragment } from '$houdini';
	import ListItem from '$lib/ui/ListItem.svelte';
	import Time from '$lib/ui/Time.svelte';
	import { BodyShort, Tooltip } from '@nais/ds-svelte-community';
	import { RocketIcon } from '@nais/ds-svelte-community/icons';
	import type { Component } from 'svelte';
	import { icons } from '../activity/activity-log-icons';
	import { activityTooltip } from '../activity/activity-log-tooltip';
	import '../activity/activity-log.css';
	import TeamCreatedActivityLogEntryText from '../activity/shared/texts/TeamCreatedActivityLogEntryText.svelte';
	import TeamUpdatedActivityLogEntryText from '../activity/shared/texts/TeamUpdatedActivityLogEntryText.svelte';
	import GroupCreatedActivityLogEntryText from '../activity/shared/texts/GroupCreatedActivityLogEntryText.svelte';
	import GroupMemberAddedActivityLogEntryText from '../activity/shared/texts/GroupMemberAddedActivityLogEntryText.svelte';
	import GroupMemberRemovedActivityLogEntryText from '../activity/shared/texts/GroupMemberRemovedActivityLogEntryText.svelte';
	import TeamRoleAssignedActivityLogEntryText from '../activity/shared/texts/TeamRoleAssignedActivityLogEntryText.svelte';
	import TeamRoleRevokedActivityLogEntryText from '../activity/shared/texts/TeamRoleRevokedActivityLogEntryText.svelte';

	interface Props {
		item: ActivityLogEntryFragment;
	}

	let { item }: Props = $props();

	let data = $derived(
		fragment(
			item,
			graphql(`
				fragment ActivityLogEntryFragment on ActivityLogEntry {
					__typename
					id
					actor
					createdAt
					message
					resourceName
					resourceType
					teamSlug
					... on TeamCreatedActivityLogEntry {
						__typename
					}
					... on TeamUpdatedActivityLogEntry {
						teamUpdated: data {
							updatedFields {
								field
								oldValue
								newValue
							}
						}
					}
					... on GroupCreatedActivityLogEntry {
						__typename
					}
					... on GroupMemberAddedActivityLogEntry {
						groupMemberAdded: data {
							userEmail
						}
					}
					... on GroupMemberRemovedActivityLogEntry {
						groupMemberRemoved: data {
							userEmail
						}
					}
					... on TeamRoleAssignedActivityLogEntry {
						roleAssigned: data {
							user {
								email
							}
							role
						}
					}
					... on TeamRoleRevokedActivityLogEntry {
						roleRevoked: data {
							user {
								email
							}
							role
						}
					}
				}
			`)
		)
	);

	const Icon = $derived(icons[$data.__typename] || RocketIcon);

	function textComponent(typename: string): Component<{ data: unknown }> | null {
		switch (typename) {
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
				return null;
		}
	}

	const TextComponent = $derived(textComponent($data.__typename));
</script>

<ListItem>
	<div style="display: flex; gap: 0.5rem;">
		<div class="activity-icon">
			<Tooltip content={activityTooltip($data.__typename)}>
				<Icon size="1em" width="1em" height="1em" />
			</Tooltip>
		</div>

		<div>
			{#if TextComponent}
				<TextComponent data={$data} />
			{:else}
				<BodyShort size="small" spacing>
					{$data.message}
				</BodyShort>
				<BodyShort size="small" style="color: var(--ax-text-subtle)">
					<Time time={$data.createdAt} distance={true} />
					by {$data.actor}
				</BodyShort>
			{/if}
		</div>
	</div>
</ListItem>
