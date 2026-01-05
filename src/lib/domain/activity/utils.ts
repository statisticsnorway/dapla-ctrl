import { ActivityLogEntryResourceType, type ActivityLogEntryResourceType$options } from '$houdini';

export const activityLogResourceLink = (
	environmentName: string,
	resourceType: ActivityLogEntryResourceType$options,
	resourceName: string,
	teamSlug: string | null
) => {
	switch (resourceType) {
		case ActivityLogEntryResourceType.TEAM:
			return `/team/${teamSlug}`;
		default:
			return null;
	}
};
