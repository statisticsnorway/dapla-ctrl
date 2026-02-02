import { load_UserTeams, OrderDirection, TeamOrderField } from '$houdini';
import { urlToOrderDirection, urlToOrderField } from '$lib/ui/DaplaTable.svelte';
import type { PageLoadEvent } from './$types';

export async function load(event: PageLoadEvent) {
	return {
		...event.data,
		...(await load_UserTeams({
			event,
			variables: {
				orderBy: {
					field: urlToOrderField(TeamOrderField, TeamOrderField.SLUG, event.url),
					direction: urlToOrderDirection(event.url, OrderDirection.ASC)
				}
			}
		}))
	};
}
