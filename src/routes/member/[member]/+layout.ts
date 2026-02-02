import { load_UserLayoutInfo } from '$houdini';
import { addPageMeta } from '$lib/utils/pageMeta.js';
import { get } from 'svelte/store';

export async function load(event) {
	const { UserLayoutInfo } = await load_UserLayoutInfo({
		event,
		blocking: true,
		variables: { user: event.params.member }
	});

	const name = get(UserLayoutInfo).data?.user.name ?? event.params.member;

	const meta = await addPageMeta(event, {
		breadcrumbs:
			event.route.id === '/member/[member]'
				? undefined
				: [
						{
							label: name,
							href: '/member/[member]'
						}
					]
	});
	return {
		...meta
	};
}
