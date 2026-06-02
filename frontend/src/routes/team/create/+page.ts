import { load_SectionsInfo } from '$houdini';

export async function load(event) {
	return {
		...(await load_SectionsInfo({
			event,
			blocking: true
		}))
	};
}
