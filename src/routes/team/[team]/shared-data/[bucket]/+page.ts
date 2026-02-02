import { load_TeamSharedBucketAccess, OrderDirection, UserOrderField } from '$houdini';
import { urlToOrderDirection, urlToOrderField } from '$lib/ui/DaplaTable.svelte';
import { addPageMeta } from '$lib/utils/pageMeta';
import type { PageLoadEvent } from './$types';
import { get } from 'svelte/store';

export async function load(event: PageLoadEvent) {
	const meta = await load_TeamSharedBucketAccess({
		event,
		variables: {
			bucket: event.params.bucket,
			orderBy: {
				field: urlToOrderField(UserOrderField, UserOrderField.NAME, event.url),
				direction: urlToOrderDirection(event.url, OrderDirection.ASC)
			}
		},
		blocking: true
	});

	const { data } = get(meta.TeamSharedBucketAccess);

	return {
		...meta,
		...event.data,
		...(await addPageMeta(event, {
			title: data?.sharedBucket.shortName ?? event.params.bucket,
			tag: {
				label: data?.sharedBucket.env ?? '',
				variant: 'info'
			}
		}))
	};
}
