import { load_AllTeams } from '$houdini';
import type { PageLoadEvent } from './$types';

export async function load(event: PageLoadEvent) {
	return {
		...(await load_AllTeams({
			event
		}))
	};
}
