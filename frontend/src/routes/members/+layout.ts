import { load_MyTeamMembers, load_AllTeamMembers } from '$houdini';
import type { LayoutLoadEvent } from './$types';

export async function load(event: LayoutLoadEvent) {
	return {
		...(await load_MyTeamMembers({
			event,
			blocking: false
		})),
		...(await load_AllTeamMembers({
			event,
			variables: {
				first: 1
			}
		}))
	};
}
