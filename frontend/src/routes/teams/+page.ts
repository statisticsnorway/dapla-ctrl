import { load_AllTeams, OrderDirection, TeamOrderField } from '$houdini';
import { urlToOrderDirection, urlToOrderField } from '$lib/ui/DaplaTable.svelte';
import type { PageLoadEvent } from './$types';

export async function load(event: PageLoadEvent) {
	return {
		...event.data,
		...(await load_AllTeams({
			event,
			blocking: true,
			variables: {
				orderBy: {
					field: urlToOrderField(TeamOrderField, TeamOrderField.SLUG, event.url),
					direction: urlToOrderDirection(event.url, OrderDirection.ASC)
				}
			}
		}))
	};
}
