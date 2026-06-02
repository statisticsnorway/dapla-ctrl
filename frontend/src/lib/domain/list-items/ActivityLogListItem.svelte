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
	import ServiceAccountCreatedActivityLogEntryText from '../activity/shared/texts/ServiceAccountCreatedActivityLogEntryText.svelte';
	import ServiceAccountTokenCreatedActivityLogText from '../activity/shared/texts/ServiceAccountTokenCreatedActivityLogEntryText.svelte';
	import TeamRoleRevokedActivityLogEntryText from '../activity/shared/texts/TeamRoleRevokedActivityLogEntryText.svelte';
	import ReconcilerConfiguredActivityLogEntryText from '../activity/shared/texts/ReconcilerConfiguredActivityLogEntryText.svelte';
	import ReconcilerDisabledActivityLogEntryText from '../activity/shared/texts/ReconcilerDisabledActivityLogEntryText.svelte';
	import ReconcilerEnabledActivityLogEntryText from '../activity/shared/texts/ReconcilerEnabledActivityLogEntryText.svelte';
	import ServiceAccountDeletedActivityLogEntryText from '../activity/shared/texts/ServiceAccountDeletedActivityLogEntryText.svelte';
	import RoleAssignedToServiceAccountActivityLogEntryText from '../activity/shared/texts/RoleAssignedToServiceAccountActivityLogEntryText.svelte';
	import RoleRevokedFromServiceAccountActivityLogEntryText from '../activity/shared/texts/RoleRevokedFromServiceAccountActivityLogEntryText.svelte';
	import ServiceAccountTokenDeletedActivityLogEntryText from '../activity/shared/texts/ServiceAccountTokenDeletedActivityLogEntryText.svelte';
	import ServiceAccountTokenUpdatedActivityLogEntryText from '../activity/shared/texts/ServiceAccountTokenUpdatedActivityLogEntryText.svelte';
	import ServiceAccountUpdatedActivityLogEntryText from '../activity/shared/texts/ServiceAccountUpdatedActivityLogEntryText.svelte';

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
						__typename
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
					... on ReconcilerConfiguredActivityLogEntry {
						__typename
						data {
							updatedKeys
						}
					}
					... on ReconcilerDisabledActivityLogEntry {
						__typename
						resourceName
					}
					... on ReconcilerEnabledActivityLogEntry {
						__typename
						resourceName
					}

					... on ServiceAccountCreatedActivityLogEntry {
						__typename
						resourceName
					}
					... on ServiceAccountDeletedActivityLogEntry {
						__typename
						resourceName
					}
					... on ServiceAccountUpdatedActivityLogEntry {
						__typename
						data {
							updatedFields {
								field
								oldValue
								newValue
							}
						}
					}

					... on ServiceAccountTokenCreatedActivityLogEntry {
						__typename
						resourceName
					}
					... on ServiceAccountTokenDeletedActivityLogEntry {
						__typename
						resourceName
						data {
							tokenName
						}
					}
					... on ServiceAccountTokenUpdatedActivityLogEntry {
						__typename
						data {
							updatedFields {
								field
								oldValue
								newValue
							}
						}
					}
					... on RoleAssignedToServiceAccountActivityLogEntry {
						__typename
						data {
							roleName
						}
					}
					... on RoleRevokedFromServiceAccountActivityLogEntry {
						__typename
						data {
							roleName
						}
					}
				}
			`)
		)
	);

	const Icon = $derived(icons[$data.__typename] || RocketIcon);

	function textComponent(typename: string): Component<{ data: unknown }> | null {
		switch (typename) {
			case 'ReconcilerConfiguredActivityLogEntry':
				return ReconcilerConfiguredActivityLogEntryText as Component<{ data: unknown }>;
			case 'ReconcilerDisabledActivityLogEntry':
				return ReconcilerDisabledActivityLogEntryText as Component<{ data: unknown }>;
			case 'ReconcilerEnabledActivityLogEntry':
				return ReconcilerEnabledActivityLogEntryText as Component<{ data: unknown }>;
			case 'ServiceAccountDeletedActivityLogEntry':
				return ServiceAccountDeletedActivityLogEntryText as Component<{ data: unknown }>;
			case 'RoleAssignedToServiceAccountActivityLogEntry':
				return RoleAssignedToServiceAccountActivityLogEntryText as Component<{ data: unknown }>;
			case 'RoleRevokedFromServiceAccountActivityLogEntry':
				return RoleRevokedFromServiceAccountActivityLogEntryText as Component<{ data: unknown }>;
			case 'ServiceAccountTokenDeletedActivityLogEntry':
				return ServiceAccountTokenDeletedActivityLogEntryText as Component<{ data: unknown }>;
			case 'ServiceAccountTokenUpdatedActivityLogEntry':
				return ServiceAccountTokenUpdatedActivityLogEntryText as Component<{ data: unknown }>;
			case 'ServiceAccountUpdatedActivityLogEntry':
				return ServiceAccountUpdatedActivityLogEntryText as Component<{ data: unknown }>;

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
			case 'ServiceAccountCreatedActivityLogEntry':
				return ServiceAccountCreatedActivityLogEntryText as Component<{ data: unknown }>;
			case 'ServiceAccountTokenCreatedActivityLogEntry':
				return ServiceAccountTokenCreatedActivityLogText as Component<{ data: unknown }>;
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
				<BodyShort size="small" textColor="subtle">
					av {$data.actor} for
					<Time time={$data.createdAt} distance />
				</BodyShort>
			{/if}
		</div>
	</div>
</ListItem>
