import { load_TeamSharedBucketAccess } from '$houdini';
import { addPageMeta } from '$lib/utils/pageMeta';
import type { PageLoadEvent } from './$types';

export async function load(event: PageLoadEvent) {
	return {
		...(await addPageMeta(event, {
			title: event.params.bucket,
			breadcrumbs: [{ label: 'Datadeling', href: '/team/[team]/shared-data' }]
		})),
		...event.data,
		...(await load_TeamSharedBucketAccess({
			event,
			variables: { bucket: event.params.bucket },
			blocking: true
		}))
	};
}
