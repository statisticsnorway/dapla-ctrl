// src/lib/components/activity/activity-log-icons.ts
import type { Component } from 'svelte';

// Resource icons (same "shapes" you use elsewhere)
import {
	CogIcon,
	HexagonGridIcon,
	PersonGroupIcon,
	RobotSmileIcon
} from '@nais/ds-svelte-community/icons';

/**
 * ICON SHAPES (what is operated on)
 * Keep this aligned with Icon.svelte's mental model:
 * - Group/Members → PersonGroupIcon
 * - Team → HexagonGridIcon
 */
export const icons: { [typename: string]: Component } = {
	/* Team & members */
	GroupMemberAddedActivityLogEntry: PersonGroupIcon,
	GroupMemberRemovedActivityLogEntry: PersonGroupIcon,
	GroupCreatedActivityLogEntry: PersonGroupIcon,
	TeamCreatedActivityLogEntry: HexagonGridIcon,
	TeamUpdatedActivityLogEntry: HexagonGridIcon,

	/* Fallback / infra ops */
	ReconcilerConfiguredActivityLogEntry: CogIcon,
	ReconcilerEnabledActivityLogEntry: CogIcon,
	ReconcilerDisabledActivityLogEntry: CogIcon,

	ServiceAccountCreatedActivityLogEntry: RobotSmileIcon,
	ServiceAccountTokenCreatedActivityLogEntry: RobotSmileIcon,
	ServiceAccountDeletedActivityLogEntry: RobotSmileIcon,
	ServiceAccountUpdatedActivityLogEntry: RobotSmileIcon,
	ServiceAccountTokenDeletedActivityLogEntry: RobotSmileIcon,
	ServiceAccountTokenUpdatedActivityLogEntry: RobotSmileIcon,
	RoleAssignedToServiceAccountActivityLogEntry: RobotSmileIcon,
	RoleRevokedFromServiceAccountActivityLogEntry: RobotSmileIcon
};
