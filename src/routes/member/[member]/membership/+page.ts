import { load_UserMemberships } from '$houdini';
import { addPageMeta } from '$lib/utils/pageMeta.js';

export async function load(event) {
	const meta = await load_UserMemberships({
		event,
		variables: { user: event.params.member },
		blocking: true
	});
	return {
		...meta,
		...event.data,
		...(await addPageMeta(event, { title: 'Medlemskap' }))
	};
}
