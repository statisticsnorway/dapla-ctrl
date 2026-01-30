import { addPageMeta } from '$lib/utils/pageMeta.js';

export async function load(event) {
	const meta = await addPageMeta(event, {
		title: 'Medlemmer',
		breadcrumbs:
			event.route.id === '/user'
				? undefined
				: [
						{
							label: 'Medlemmer',
							href: `/team-members`
						}
					]
	});
	return {
		...meta
	};
}
