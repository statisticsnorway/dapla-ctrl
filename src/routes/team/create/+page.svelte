<script lang="ts">
	import { enhance } from '$app/forms';
	import { isPossiblyInModal } from '$lib/ui/PageModal.svelte';
	import WarningIcon from '$lib/icons/WarningIcon.svelte';
	import { Button, ErrorSummary, Heading, TextField } from '@nais/ds-svelte-community';
	import { FloppydiskIcon } from '@nais/ds-svelte-community/icons';
	import type { PageProps } from './$types';

	let { form, data }: PageProps = $props();
	let saving = $state(false);

	let { UserInfo } = $derived(data);
	const isAdmin = $derived(
		$UserInfo.data?.me.__typename === 'User' ? $UserInfo.data?.me.isAdmin : false
	);

	let teamSlugError = $state('');

	let purposeError = $state('');

	let sectionError = $state('no_error');

	let disabled = $derived(
		[teamSlugError, purposeError, sectionError].some((e) => e !== 'no_error')
	);

	const reservedSlugs = [
		'kube-system',
		'kube-node-lease',
		'kube-public',
		'kyverno',
		'cnrm-system',
		'configconnector-operator-system',
		'default'
	];
	const slugPattern = /^[a-z][a-z0-9-]{0,15}[a-z]$$/;

	function handleTeamSlugInput(event: Event) {
		if (!event) return;
		const input = event.target as HTMLInputElement | null;
		if (input) {
			const slug = input.value;

			// Check if the slug is reserved
			if (reservedSlugs.includes(slug)) {
				teamSlugError = 'Dette navnet reservert.';
				return;
			}

			// Check if the slug starts with "team"
			if (slug.startsWith('team')) {
				teamSlugError =
					"Pefikset 'team' er overflødig. Når du oppretter et team, er det per definisjon et team. Prøv igjen med et annet navn, kanskje bare ved å fjerne prefikset?";
				return;
			}

			// Check the length of the slug
			if (slug.length < 3) {
				teamSlugError = 'Navnet må være minst 3 tegn langt.';
				return;
			}

			if (slug.length > 17) {
				teamSlugError = 'Navnet kan maksimalt være 17 tegn langt.';
				return;
			}

			// Validate the slug against the pattern
			if (!slugPattern.test(slug)) {
				teamSlugError =
					'Navnet kan kun inneholde små bokstaver, tall og bindestreker. Det må starte og slutte med små bokstaver.';
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

	const sectionPattern = /^([0-9]{3})?$/;
	function handleSectionInput(event: Event) {
		if (!event) return;
		const input = event.target as HTMLInputElement | null;
		if (input) {
			if (!sectionPattern.test(input.value)) {
				sectionError = 'Seksjonskoden må være tom eller 3 siffer.';
			} else {
				sectionError = 'no_error';
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
				Teknisk navn
			{/snippet}
			{#snippet description()}
				Eksempel: mitt-team-navn<br />
				<WarningIcon class="text-aligned-icon" /> Det er ikke mulig å endre teknisk etter opprettelse,
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
		{#if isAdmin}
			<br />
			<TextField name="section" value={form?.input.sectionCode} oninput={handleSectionInput}>
				{#snippet label()}
					Seksjonskode
				{/snippet}
				{#snippet description()}
					Eksempel: 724
				{/snippet}
			</TextField>
			{#if sectionError !== 'no_error' && sectionError !== ''}
				<p style:color="var(--ax-text-danger)">{sectionError}</p>
			{/if}
		{/if}
		<br />
		<Button loading={saving} {disabled} icon={FloppydiskIcon}>Opprett team</Button>
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
