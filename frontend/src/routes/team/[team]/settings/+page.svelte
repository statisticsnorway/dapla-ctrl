<script lang="ts">
	import { graphql } from '$houdini';
	import GraphErrors from '$lib/ui/GraphErrors.svelte';
	import { Heading, Switch } from '@nais/ds-svelte-community';
	import type { PageProps } from './$types';
	import Confirm from '$lib/ui/Confirm.svelte';

	let { data }: PageProps = $props();
	let { TeamSettings, teamSlug } = $derived(data);
	let showConfirmModal = $state(false);
	let aiErrors: { message: string }[] | undefined = $state();

	const updateTeam = graphql(`
		mutation UpdateTeam($input: UpdateTeamInput!) {
			updateTeam(input: $input) {
				team {
					hasManualEditing
				}
			}
		}
	`);
	const enableTeamFeature = graphql(`
		mutation EnableTeamFeature($input: EnableTeamFeatureInput!) {
			enableTeamFeature(input: $input) {
				team {
					features {
						name
						env
					}
				}
			}
		}
	`);
	const disableTeamFeature = graphql(`
		mutation DisableTeamFeature($input: DisableTeamFeatureInput!) {
			disableTeamFeature(input: $input) {
				team {
					features {
						name
						env
					}
				}
			}
		}
	`);

	let teamSettings = $derived($TeamSettings.data?.team);
	let aiEnabled = $derived(
		teamSettings?.features.some(({ name, env }) => name === 'ai' && env === 'test')
	);

	let descriptionErrors: { message: string }[] | undefined = $state();

	const toggle = async () => {
		descriptionErrors = undefined;
		const data = await updateTeam.mutate({
			input: {
				slug: teamSlug,
				hasManualEditing: !$TeamSettings.data?.team.hasManualEditing
			}
		});

		if (data.errors) {
			descriptionErrors = data.errors;
		}
	};

	const toggleAi = async () => {
		aiErrors = undefined;
		const result = await (aiEnabled ? disableTeamFeature : enableTeamFeature).mutate({
			input: { teamSlug, feature: 'ai', env: 'test' }
		});
		if (result.errors) aiErrors = result.errors;
	};
</script>

<GraphErrors errors={$TeamSettings.errors} />

{#if teamSettings}
	<div class="wrapper">
		<div style="display: flex; flex-direction: column; gap: var(--spacing-layout)">
			<div>
				<Heading level="2">Parquedit</Heading>

				Parquedit er en lagringsløsning for manuell editering, levert av team Dapla
				Fellesfunksjoner.

				<Switch
					checked={teamSettings.hasManualEditing}
					onclick={(e: MouseEvent) => {
						e.preventDefault();
						showConfirmModal = true;
					}}
					>{teamSettings.hasManualEditing
						? 'Fjern tilgang til Parquedit'
						: 'Aktiver tilgang til Parquedit'}</Switch
				>

				<GraphErrors errors={descriptionErrors} size="small" />
			</div>
			<div>
				<Heading level="2">Kunsting Intelligens</Heading>
				Aktiver KI-funksjonalitet for teamet i testmiljøet.
				<Switch checked={aiEnabled} onclick={toggleAi}>
					{aiEnabled ? 'Deaktiver KI' : 'Aktiver KI'}
				</Switch>
				<GraphErrors errors={aiErrors} size="small" />
			</div>
		</div>
	</div>
{/if}

<Confirm bind:open={showConfirmModal} onconfirm={toggle} confirmText="Bekreft">
	{#snippet header()}
		<h3>Parquedit bekreftelse</h3>
	{/snippet}

	Dette vil <b>{$TeamSettings.data?.team.hasManualEditing ? 'slå av' : 'slå på'}</b> Parquedit.<br
	/>
	{$TeamSettings.data?.team.hasManualEditing
		? 'Historikken knyttet til manuell editering vil da bli slettet.'
		: ''}
</Confirm>

<style>
	.wrapper {
		display: grid;
		grid-template-columns: 1fr 320px;
		gap: var(--spacing-layout);
	}
</style>
