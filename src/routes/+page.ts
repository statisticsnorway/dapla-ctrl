import { load_UserTeams } from '$houdini';
import type { PageLoadEvent } from './$types';

export async function load(event: PageLoadEvent) {
	return {
		...(await load_UserTeams({
			event
		}))
	};
}
