import { load_UserOverview } from '$houdini';
import { addPageMeta } from '$lib/utils/pageMeta.js';
import { get } from 'svelte/store';

export async function load(event) {
	const meta = await load_UserOverview({
		event,
		variables: { user: event.params.user },
		blocking: true
	});
	const { data } = get(meta.UserOverview);
	return {
		...meta,
		...event.data,
		...(await addPageMeta(event, { title: data?.user.name ?? 'Ukjent bruker' }))
	};
}
