import { load_AllTeams } from '$houdini';

export async function load(event) {
	return {
		...(await load_AllTeams({
			event
		}))
	};
}
