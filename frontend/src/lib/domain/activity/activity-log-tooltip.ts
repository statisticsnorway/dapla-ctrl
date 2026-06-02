// src/lib/components/activity/activity-log-tooltip.ts

/**
 * Returns a user-friendly tooltip label for an activity log entry type.
 */
export function activityTooltip(typename: string): string {
	switch (typename) {
		case 'TeamCreatedActivityLogEntry':
		case 'TeamUpdatedActivityLogEntry':
			return 'Team';
		default:
			return 'Activity';
	}
}
