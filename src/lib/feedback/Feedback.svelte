<script lang="ts">
	import { page } from '$app/state';
	import { replacer } from '$lib/replacer';
	import { themeSwitch } from '$lib/stores/theme.svelte';
	import { Button, Checkbox, Heading, Modal, Select, Theme } from '@nais/ds-svelte-community';
	import type { FeedbackType } from './types';

	interface Props {
		close: () => void;
	}

	let { close }: Props = $props();

	let type: FeedbackType | '' = $state('');
	let details = $state('');
	let anonymous: boolean = $state(false);
	let uri = '';

	let loading = $state(false);
	let feedbackSent: boolean = $state(false);

	let errorMessage: string = $state('');
	let errorType: boolean = $state(false);
	let errorDetails: boolean = $state(false);

	const maxlength = 3000;

	if (page.route.id !== null) {
		uri = replacer(page.route.id, page.params);
	}

	const FEEDBACK_TYPE: { value: FeedbackType | ''; text: string }[] = [
		{ value: '', text: 'Velg tilbakemeldingstype' },
		{ value: 'KUDOS', text: 'Skryt' },
		{ value: 'BUG', text: 'Feil' },
		{ value: 'CHANGE_REQUEST', text: 'Funksjonalitets- eller endringsønske' },
		{ value: 'QUESTION', text: 'Spørsmål' },
		{ value: 'OTHER', text: 'Annet' }
	];

	const submitFeedback = async () => {
		if (type === '') {
			errorType = true;
		} else {
			errorType = false;
		}

		if (details === '' || details === undefined) {
			errorDetails = true;
		} else {
			errorDetails = false;
		}

		if (errorType || errorDetails) {
			return;
		}

		loading = true;
		try {
			const result = await fetch('/api/send-feedback', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					feedback: details,
					type: type,
					path: uri,
					anonymous: anonymous
				})
			});
			loading = false;
			const data = await result.json();

			if (data.error) {
				errorMessage = data.error;
				return;
			}

			feedbackSent = true;
			return data.ok ? 'Melding sendt!' : 'Klarte ikke sende melding.';
		} catch (error) {
			console.error('Error:', error);
			return 'Det oppsto en feil under sending av melding: ' + error;
		}
	};
</script>

<Theme theme={themeSwitch.theme}>
	<Modal open width="medium" onclose={close}>
		{#snippet header()}
			<Heading level="1">Dapla Ctrl tilbakemelding</Heading>
		{/snippet}

		{#if feedbackSent}
			<p>Takk for tilbakemeldingen!</p>
		{:else}
			<p>
				Tilbakemeldingen vil bli knyttet til din e-postadresse. Huk av boksen under hvis du ønsker å
				gi tilbakemeldingen anonymt.
				<br />
				Vi ser frem til å høre fra deg!
			</p>
			<div class="wrapper">
				<Select size="small" label="Type" bind:value={type}>
					{#each FEEDBACK_TYPE as option (option)}
						<option value={option.value}>{option.text}</option>
					{/each}
				</Select>
				{#if errorType}
					<p class="aksel-error-message aksel-label aksel-label--small">
						Tilbakemeldingstype må være valgt
					</p>
				{/if}

				<label class="aksel-form-field__label aksel-label aksel-label--small" for="details">
					Beskrivelse
				</label>
				<div class="details">
					<textarea
						class="aksel-textarea__input aksel-body-short aksel-body-short--small textarea"
						id="details"
						bind:value={details}
						rows="5"
						cols="40"
						{maxlength}
						style="resize: vertical; min-height: 16rem; "
						placeholder="Skriv tilbakemeldingen din her..."
						disabled={feedbackSent}
					></textarea>
					<span id="charCount">{maxlength - details.length} tegn gjenstår</span>
				</div>
				<div
					class="aksel-form-field__error"
					id="tf-uid-43"
					aria-relevant="additions removals"
					aria-live="polite"
				>
					{#if errorDetails}
						<p class="aksel-error-message aksel-label aksel-label--small">
							Beskrivelse må fylles ut
						</p>
					{/if}
				</div>
				{#if errorMessage !== ''}
					<p class="aksel-error-message aksel-label aksel-label--small">{errorMessage}</p>
				{/if}
				<Checkbox bind:checked={anonymous}>Anonym tilbakemelding</Checkbox>
			</div>
		{/if}
		{#snippet footer()}
			{#if feedbackSent}
				<Button variant="primary" size="small" onclick={close}>Lukk</Button>
			{:else}
				<Button variant="primary" size="small" {loading} onclick={submitFeedback}>Send</Button>
				<Button variant="secondary" size="small" onclick={close}>Lukk</Button>
			{/if}
		{/snippet}
	</Modal>
</Theme>

<style>
	.wrapper {
		display: flex;
		flex-direction: column;
		padding: 1rem;
		gap: 1rem;
	}
	.details {
		display: flex;
		flex-direction: column;
		align-items: end;
	}
	#charCount {
		font-size: 0.75rem;
		color: var(--ax-text-subtle);
		margin: 0;
		padding-top: 0.2rem;
	}
</style>
