import { load_TeamOverview } from '$houdini';
import { addPageMeta } from '$lib/utils/pageMeta';
import { error } from '@sveltejs/kit';
import { get } from 'svelte/store';

export async function load(event) {
	const team = await load_TeamOverview({
		event,
		variables: { team: event.params.team },
		blocking: true
	});

	if (!team) {
		error(501, 'Noe gikk galt under innlasting');
	}

	const current = get(team.TeamOverview);
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

	return {
		...(await addPageMeta(event, {
			title: current.data?.team.displayName
		})),
		...(current.data
			? current.data.team
			: {
					displayName: '',
					slug: '',
					isManaged: false,
					section: {
						code: '',
						name: '',
						manager: null
					},
					activityLog: { nodes: [] }
				})
	};
}
