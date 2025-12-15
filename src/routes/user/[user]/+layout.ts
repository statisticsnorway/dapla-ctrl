import { addPageMeta } from '$lib/utils/pageMeta.js';

export async function load(event) {
	const meta = await addPageMeta(event, {
		breadcrumbs:
			event.route.id === '/user/[user]'
				? undefined
				: [
						{
							label: event.params.user,
							href: '/user/[user]'
						}
					]
	});
	return {
		...meta
	};
}
