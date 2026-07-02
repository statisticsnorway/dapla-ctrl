import { load_LaunchLab } from '$houdini';
import { addPageMeta } from '$lib/utils/pageMeta';
import { error } from '@sveltejs/kit';
import type { BeforeLoadEvent } from './$houdini';

export async function _houdini_beforeLoad({ parent }: BeforeLoadEvent) {
	const pd = await parent();

	if (!pd.viewerIsMember) {
		error(403, 'Du har ikke tilgang til denne siden');
	}
}

export async function load(event) {
	return {
		...(await addPageMeta(event, { title: 'Dapla Lab' })),
		...(await load_LaunchLab({
			event,
			variables: {
				team: event.params.team
			}
		}))
	};
}
