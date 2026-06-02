import { load_AllSharedData, OrderDirection, SharedBucketOrderField } from '$houdini';
import { urlToOrderDirection, urlToOrderField } from '$lib/ui/DaplaTable.svelte';
import { addPageMeta } from '$lib/utils/pageMeta';
import type { PageLoadEvent } from './$types';

export async function load(event: PageLoadEvent) {
	return {
		...(await addPageMeta(event, { title: 'Datadeling' })),
		...event.data,
		...(await load_AllSharedData({
			event,
			variables: {
				orderBy: {
					field: urlToOrderField(SharedBucketOrderField, SharedBucketOrderField.NAME, event.url),
					direction: urlToOrderDirection(event.url, OrderDirection.ASC)
				}
			},
			blocking: true
		}))
	};
}
