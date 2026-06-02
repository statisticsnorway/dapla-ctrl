import { load_TeamSharedBucketAccessTeams } from '$houdini';
import { addPageMeta } from '$lib/utils/pageMeta';
import type { PageLoadEvent } from './$types';
import { get } from 'svelte/store';

export async function load(event: PageLoadEvent) {
	const meta = await load_TeamSharedBucketAccessTeams({
		event,
		variables: { bucket: event.params.bucket },
		blocking: true
	});

	const { data } = get(meta.TeamSharedBucketAccessTeams);

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
