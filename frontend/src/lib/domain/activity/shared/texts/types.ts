import type { ActivityLogEntryFragment$data } from '$houdini';

/**
 * Union type for activity log entries from different sources.
 * This allows components to work with both list view and team overview data.
 */
type TypeKeyedProps = {
	[K in keyof ActivityLogEntryFragment$data as K extends `${string}ActivityLogEntry`
		? K
		: never]: ActivityLogEntryFragment$data[K];
};

type TypeKeys = keyof TypeKeyedProps;

export type ActivityLogEntry<T extends TypeKeys> = Omit<ActivityLogEntryFragment$data, TypeKeys> &
	NonNullable<ActivityLogEntryFragment$data[T]>;
