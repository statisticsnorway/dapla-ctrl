<script lang="ts">
	import { enhance } from '$app/forms';
	import { isPossiblyInModal } from '$lib/ui/PageModal.svelte';
	import WarningIcon from '$lib/icons/WarningIcon.svelte';
	import { Button, ErrorSummary, Heading, Select, TextField } from '@nais/ds-svelte-community';
	import { FloppydiskIcon } from '@nais/ds-svelte-community/icons';
	import type { PageProps } from './$types';

	let { form, data }: PageProps = $props();
	let saving = $state(false);

	let { UserInfo, SectionsInfo } = $derived(data);
	const isAdmin = $derived(
		$UserInfo.data?.me.__typename === 'User' ? $UserInfo.data?.me.isAdmin : false
	);

	const isSection7xxManager = $derived(
		$UserInfo.data?.me.__typename === 'User'
			? $UserInfo.data?.me.isSectionManager && $UserInfo.data?.me.section?.code?.startsWith('7')
			: false
	);

	let displayNameError = $state('');

	let teamSlugError = $state('');

	let purposeError = $state('');

	let sectionError = $state('no_error');

	let sections = $derived(
		$SectionsInfo.data?.sections.nodes.toSorted((a, b) => (a.code < b.code ? -1 : 1)) ?? []
	);

	let disabled = $derived(
		[displayNameError, teamSlugError, purposeError, sectionError].some((e) => e !== 'no_error')
	);

	function handleDisplayNameInput(event: Event) {
		if (!event) return;
		const input = event.target as HTMLInputElement | null;
		if (input) {
			// TODO: Add validation rules
			displayNameError = 'no_error';
		}
	}

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
				teamSlugError = 'Dette navnet er opptatt.';
				return;
			}

			// Check if the slug starts with "team"
			if (slug.startsWith('team')) {
				teamSlugError =
					"Pefikset 'team' er overflødig. Når du oppretter et team, er det per definisjon et team. Prøv igjen med et annet navn, kanskje bare ved å fjerne prefikset?";
				return;
			}

			// Check the length of the slug
			if (slug.length < 2) {
				teamSlugError = 'Teknisk navn må være minst 2 tegn langt.';
				return;
			}

			if (slug.length > 17) {
				teamSlugError = 'Teknisk navn kan maksimalt være 17 tegn langt.';
				return;
			}

			// Validate the slug against the pattern
			if (!slugPattern.test(slug)) {
				teamSlugError =
					'Teknisk navn kan kun inneholde små bokstaver, tall og bindestreker. Det må starte og slutte med små bokstaver.';
				return;
			}

			//validate team name doesn't contain anything related to groups
			if (['managers', 'consumers', 'data-admins', 'developers'].some((el) => slug.includes(el))) {
				teamSlugError =
					'Teknisk navn kan ikke inneholde gruppenavn. Grupper defineres etter teamet er opprettet.';
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
		Se "<a
			href="https://manual.dapla.ssb.no/statistikkere/hva-er-dapla-team.html#navnestruktur"
			target="_blank">Navnestruktur</a
		>" på Dapla-manualen for hvordan finne et godt teamnavn.
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
		<TextField name="displayname" value={form?.input.displayName} oninput={handleDisplayNameInput}>
			{#snippet label()}
				Visningsnavn
			{/snippet}
			{#snippet description()}
				Eksempel: Mitt teamnavn<br />
			{/snippet}
		</TextField>
		{#if displayNameError !== 'no_error' && displayNameError !== ''}
			<p style:color="var(--ax-text-danger)">{displayNameError}</p>
		{/if}
		<br />
		<TextField name="name" value={form?.input.slug} oninput={handleTeamSlugInput}>
			{#snippet label()}
				Teknisk navn
			{/snippet}
			{#snippet description()}
				Eksempel: mitt-team-navn<br />
				<WarningIcon class="text-aligned-icon" /> Det er ikke mulig å endre teknisk navn etter opprettelse.
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
		{#if isAdmin || isSection7xxManager}
			<br />
			<Select name="isManaged" label="Autonomitetsnivå" value={form?.input.isManaged}>
				<option value="">Managed</option>
				<option value="false">Self-managed</option>
			</Select>
		{/if}
		{#if isAdmin}
			<br />
			<Select name="section" label="Eierseksjon" value={form?.input.sectionCode}>
				<option value=""></option>
				{#each sections as section (section.code)}
					<option value={section.code}>{section.code} {section.name}</option>
				{/each}
				{#snippet description()}
					Seksjonen teamet blir opprettet i. Tomt felt oppretter teamet i seksjon 724.
				{/snippet}
			</Select>
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
