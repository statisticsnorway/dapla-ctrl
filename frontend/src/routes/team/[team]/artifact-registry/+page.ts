import { load_ArtifactRegistryGithubRepos } from '$houdini';
import { addPageMeta } from '$lib/utils/pageMeta';

export async function load(event) {
	return {
		...(await addPageMeta(event, { title: 'Artifact Registry' })),
		...(await load_ArtifactRegistryGithubRepos({
			event,
			blocking: true,
			variables: { team: event.params.team }
		}))
	};
}
