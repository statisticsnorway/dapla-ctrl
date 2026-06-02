<script lang="ts">
	import { graphql } from '$houdini';
	import GraphErrors from '$lib/ui/GraphErrors.svelte';
	import { Heading } from '@nais/ds-svelte-community';
	import type { PageProps } from './$types';
	import EditText from './EditText.svelte';

	let { data }: PageProps = $props();
	let { TeamSettings, teamSlug, viewerIsMember } = $derived(data);

	const updateTeam = graphql(`
		mutation UpdateTeam($input: UpdateTeamInput!) {
			updateTeam(input: $input) {
				team {
					displayName
				}
			}
		}
	`);

	let teamSettings = $derived($TeamSettings.data?.team);

	let descriptionErrors: { message: string }[] | undefined = $state();
</script>

<GraphErrors errors={$TeamSettings.errors} />

{#if teamSettings}
	<div class="wrapper">
		<div style="display: flex; flex-direction: column; gap: var(--spacing-layout)">
			<div>
				<Heading level="2">Visningsnavn</Heading>
				<EditText
					text={teamSettings.displayName}
					onsave={async (text) => {
						descriptionErrors = undefined;
						const data = await updateTeam.mutate({
							input: {
								slug: teamSlug,
								displayName: text
							}
						});

						if (data.errors) {
							descriptionErrors = data.errors;
						}
					}}
					isMember={viewerIsMember}
				/>

				<GraphErrors errors={descriptionErrors} size="small" />
			</div>
		</div>
	</div>
{/if}

<style>
	.wrapper {
		display: grid;
		grid-template-columns: 1fr 320px;
		gap: var(--spacing-layout);
	}
</style>
