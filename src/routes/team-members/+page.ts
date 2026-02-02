import { load_MyTeamMembers, OrderDirection, UserOrderField } from '$houdini';
import { urlToOrderDirection, urlToOrderField } from '$lib/ui/DaplaTable.svelte';
import type { PageLoadEvent } from './$types';

const rows = 25;

export async function load(event: PageLoadEvent) {
	const after = event.url.searchParams.get('after') || '';
	const before = event.url.searchParams.get('before') || '';
	return {
		...event.data,
		...(await load_MyTeamMembers({
			event,
			variables: {
				orderBy: {
					field: urlToOrderField(UserOrderField, UserOrderField.NAME, event.url),
					direction: urlToOrderDirection(event.url, OrderDirection.ASC)
				},
				...(before ? { before, last: rows } : { after, first: rows })
			}
		}))
	};
}
