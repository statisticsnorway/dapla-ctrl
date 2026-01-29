import { load_ConsumesSharedData } from '$houdini';
import { addPageMeta } from '$lib/utils/pageMeta';
import type { PageLoadEvent } from './$types';

export async function load(event: PageLoadEvent) {
	return {
		...(await addPageMeta(event, { title: 'Datadeling' })),
		...event.data,
		...(await load_ConsumesSharedData({
			event,
			variables: { team: event.params.team },
			blocking: true
		}))
	};
}
