import { load_LaunchLab } from '$houdini';
import { addPageMeta } from '$lib/utils/pageMeta';

export async function load(event) {
	return {
		...(await addPageMeta(event, {
			title: 'Dapla Lab',
			tag: {
				label: 'Eksperimentell',
				variant: 'warning-moderate',
				tooltip: 'Denne siden er eksperimentell og kan bli fjernet. Gi gjerne tilbakemelding!'
			}
		})),
		...(await load_LaunchLab({
			event,
			variables: {
				team: event.params.team
			},
			blocking: true
		}))
	};
}
