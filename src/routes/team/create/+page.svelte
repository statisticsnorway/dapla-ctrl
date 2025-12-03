<script lang="ts">
	import { enhance } from '$app/forms';
	import { isPossiblyInModal } from '$lib/ui/PageModal.svelte';
	import WarningIcon from '$lib/icons/WarningIcon.svelte';
	import { Button, ErrorSummary, Heading, TextField } from '@nais/ds-svelte-community';
	import { FloppydiskIcon } from '@nais/ds-svelte-community/icons';
	import type { PageProps } from './$types';

	let { form }: PageProps = $props();
	let saving = $state(false);

	let teamSlugError = $state('');

	let purposeError = $state('');

	let disabled = $derived(teamSlugError !== 'no_error' || purposeError !== 'no_error');

	const reservedSlugs = [
		'kube-system',
		'kube-node-lease',
		'kube-public',
		'kyverno',
		'cnrm-system',
		'configconnector-operator-system',
		'default'
	];
	const slugPattern = /^[a-z](-?[a-z0-9]+)+$/;

	function handleTeamSlugInput(event: Event) {
		if (!event) return;
		const input = event.target as HTMLInputElement | null;
		if (input) {
			const slug = input.value;

			// Check if the slug is reserved
			if (reservedSlugs.includes(slug)) {
				teamSlugError = 'Denne slug-en er reservert.';
				return;
			}

			// Check if the slug starts with "team"
			if (slug.startsWith('team')) {
				teamSlugError =
					"Navneprefikset 'team' er overflødig. Når du oppretter et team, er det per definisjon et team. Prøv igjen med et annet navn, kanskje bare ved å fjerne prefikset?";
				return;
			}

			// Check the length of the slug
			if (slug.length < 3) {
				teamSlugError = 'En team-slug må være minst 3 tegn lang.';
				return;
			}

			if (slug.length > 30) {
				teamSlugError = 'En team-slug må være maksimalt 30 tegn lang.';
				return;
			}

			// Validate the slug against the pattern
			if (!slugPattern.test(slug)) {
				teamSlugError =
					'En team-slug må begynne med en liten bokstav og kan inneholde små bokstaver, tall og bindestreker. Den kan imidlertid ikke starte eller slutte med en bindestrek, og kan ikke inneholde påfølgende bindestreker.';
				return;
			}

			// If all validations pass, clear the error
			teamSlugError = 'no_error';
		}
	}

	function handlePurposeInput(event: Event) {
		if (!event) return;
		const input = event.target as HTMLInputElement | null;
		if (input) {
			if (input.value.length < 3) {
				purposeError = 'Formålet må være minst 3 tegn langt.';
			} else {
				purposeError = 'no_error';
			}
		}
	}
</script>

<svelte:head>
	<title>Opprett et nytt team - Dapla Ctrl</title>
</svelte:head>

<div class="container" class:partOfModal={isPossiblyInModal()}>
	{#if !isPossiblyInModal()}
		<Heading level="1" size="large" spacing>Opprett et nytt team</Heading>
	{/if}
	{#if form?.errors && form.errors.length > 0}
		<ErrorSummary heading="Feil ved opprettelse av team">
			{#each form.errors as error (error)}
				<li style="color:inherit!important">{error.message}</li>
			{/each}
		</ErrorSummary>
	{/if}
	<p>
		Å opprette et team i Nais gir tilgang til visse Nais-funksjoner, som Google Cloud-prosjekter,
		Kubernetes-namespacer eller ditt eget GitHub-team. Etter at teamet er opprettet, blir du
		administrator for teamet, med rettigheter til å legge til og fjerne teammedlemmer.
		Identifikatoren er den primære nøkkelen og vil bli brukt på tvers av systemer slik at de enkelt
		kan gjenkjennes.
	</p>
	<form
		method="POST"
		use:enhance={() => {
			saving = true;
			return async ({ update }) => {
				saving = false;
				update({ reset: false });
			};
		}}
	>
		<TextField name="name" value={form?.input.slug} oninput={handleTeamSlugInput}>
			{#snippet label()}
				Identifikator / Navn
			{/snippet}
			{#snippet description()}
				Eksempel: mitt-team-navn<br />
				<WarningIcon class="text-aligned-icon" /> Det er ikke mulig å endre identifikatoren etter opprettelse,
				så velg klokt.
			{/snippet}
		</TextField>
		{#if teamSlugError !== 'no_error' && teamSlugError !== ''}
			<p style:color="var(--ax-text-danger)">{teamSlugError}</p>
		{/if}
		<br />
		<TextField name="description" value={form?.input.purpose} oninput={handlePurposeInput}>
			{#snippet label()}
				Teamets formål
			{/snippet}
			{#snippet description()}
				Eksempel: Sørger for at brukere får en god opplevelse
			{/snippet}
		</TextField>
		{#if purposeError !== 'no_error' && purposeError !== ''}
			<p style:color="var(--ax-text-danger)">{purposeError}</p>
		{/if}
		<br />
		<Button loading={saving} {disabled} icon={FloppydiskIcon}>Opprette team</Button>
	</form>
</div>

<style>
	.container {
		padding-top: 4rem;
		margin-inline: auto;
		max-width: 620px;

		&.partOfModal {
			padding-top: 0;
		}
	}
</style>
