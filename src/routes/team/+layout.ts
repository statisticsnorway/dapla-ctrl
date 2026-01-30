import { addPageMeta } from '$lib/utils/pageMeta.js';

export async function load(event) {
	const meta = await addPageMeta(event, {
		breadcrumbs: [
			{
				label: 'Team',
				href: `/`
			}
		]
	});
	return {
		...meta
	};
}
