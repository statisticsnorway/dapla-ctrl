import { load_SharedData, OrderDirection, SharedBucketOrderField } from '$houdini';
import { urlToOrderDirection, urlToOrderField } from '$lib/ui/DaplaTable.svelte';
import { addPageMeta } from '$lib/utils/pageMeta';
import type { PageLoadEvent } from './$types';

export async function load(event: PageLoadEvent) {
	return {
		...(await addPageMeta(event, { title: 'Datadeling' })),
		...event.data,
		...(await load_SharedData({
			event,
			variables: {
				team: event.params.team,
				orderBy: {
					field: urlToOrderField(SharedBucketOrderField, SharedBucketOrderField.NAME, event.url),
					direction: urlToOrderDirection(event.url, OrderDirection.ASC)
				}
			},
			blocking: true
		}))
	};
}
