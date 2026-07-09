import type { ActivityLogEntryFragment$data, TeamOverviewActivityLog$result } from '$houdini';

/**
 * Union type for activity log entries from different sources.
 * This allows components to work with both list view and team overview data.
 */
type TypeKeyedPropsActivityLog = {
	[K in keyof ActivityLogEntryFragment$data as K extends `${string}ActivityLogEntry`
		? K
		: never]: ActivityLogEntryFragment$data[K];
};

type TeamOverviewActivityLogEntry =
	TeamOverviewActivityLog$result['team']['activityLog']['edges'][number]['node'];
type TypeKeyedPropsTeamOverviewLog = {
	[K in keyof TeamOverviewActivityLogEntry as K extends `${string}ActivityLogEntry`
		? K
		: never]: TeamOverviewActivityLogEntry[K];
};

type TypeKeys = keyof TypeKeyedPropsActivityLog | keyof TypeKeyedPropsTeamOverviewLog;

export type ActivityLogEntry<T extends TypeKeys> = Omit<ActivityLogEntryFragment$data, TypeKeys> &
	NonNullable<ActivityLogEntryFragment$data[T]>;
