import { load_UserSharedBucketAccess } from '$houdini';
import { addPageMeta } from '$lib/utils/pageMeta';
import type { PageLoadEvent } from './$types';

export async function load(event: PageLoadEvent) {
	return {
		...(await addPageMeta(event, {
			title: 'Datadeling'
		})),
		...event.data,
		...(await load_UserSharedBucketAccess({
			event,
			variables: { user: event.params.user },
			blocking: true
		}))
	};
}
