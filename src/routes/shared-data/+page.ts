import { load_AllSharedData } from '$houdini';
import { addPageMeta } from '$lib/utils/pageMeta';
import type { PageLoadEvent } from './$types';

export async function load(event: PageLoadEvent) {
	return {
		...(await addPageMeta(event, { title: 'Delte data' })),
		...event.data,
		...(await load_AllSharedData({
			event,
			blocking: true
		}))
	};
}
