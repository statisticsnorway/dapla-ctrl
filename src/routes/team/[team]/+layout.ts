import { load_TeamRoles } from '$houdini';
import { addPageMeta } from '$lib/utils/pageMeta.js';
import { error } from '@sveltejs/kit';
import { get } from 'svelte/store';

export async function load(event) {
	const roles = await load_TeamRoles({
		event,
		variables: { team: event.params.team },
		blocking: true
	});

	if (!roles) {
		error(501, 'Noe gikk galt under innlasting');
	}

	const current = get(roles.TeamRoles);
	const team = get(roles.TeamRoles).data?.team.displayName ?? event.params.team;
	if (!current) {
		error(404, 'Fant ikke teamet');
	}

	if (current.errors) {
		if (current.errors) {
			if (current.errors[0].message === 'The specified team was not found.') {
				error(404, 'Fant ikke teamet');
			}
		}
		error(500, 'Noe gikk galt under innlasting');
	}

	const meta = await addPageMeta(event, {
		breadcrumbs:
			event.route.id === '/team/[team]'
				? undefined
				: [
						{
							label: team,
							href: `/team/[team]`
						}
					]
	});

	return {
		...meta,
		...(current.data
			? current.data.team
			: {
					viewerIsOwner: false,
					deletionInProgress: false,
					lastSuccessfulSync: null,
					viewerIsMember: false,
					externalResources: { gitHubTeam: null },
					displayName: '',
					isManaged: true,
					slackChannel: '',
					members: { pageInfo: { totalCount: 0 } }
				}),
		teamSlug: event.params.team
	};
}
