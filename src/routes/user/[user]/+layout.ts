import { load_UserLayoutInfo } from '$houdini';
import { addPageMeta } from '$lib/utils/pageMeta.js';
import { get } from 'svelte/store';

export async function load(event) {
	const { UserLayoutInfo } = await load_UserLayoutInfo({
		event,
		blocking: true,
		variables: { user: event.params.user }
	});

	const name = get(UserLayoutInfo).data?.user.name ?? event.params.user;

	const meta = await addPageMeta(event, {
		breadcrumbs:
			event.route.id === '/user/[user]'
				? undefined
				: [
						{
							label: name,
							href: '/user/[user]'
						}
					]
	});
	return {
		...meta
	};
}
