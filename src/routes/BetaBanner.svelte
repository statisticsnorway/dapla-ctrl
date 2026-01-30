<script lang="ts">
	import Feedback from '$lib/feedback/Feedback.svelte';
	import { Button } from '@nais/ds-svelte-community';
	import { ChatElipsisIcon, InformationIcon } from '@nais/ds-svelte-community/icons';

	let feedbackOpen = $state(false);
</script>

<div class="beta-message-container">
	<p class="beta-message">
		<InformationIcon style="width: 1.5rem;height: 1.5rem; vertical-align: bottom;" /> Velkommen til testing
		av nye Dapla Ctrl. Det er ikke mulig å utføre handlinger her enda, men dataen er reell. Gi oss gjerne
		<Button
			variant="tertiary-neutral"
			icon={ChatElipsisIcon}
			size="xsmall"
			onclick={() => {
				feedbackOpen = true;
			}}
			iconPosition="right"
		>
			<span style="font-weight: 400">tilbakemelding</span>
		</Button>
	</p>
	{#if feedbackOpen}
		<Feedback close={() => (feedbackOpen = false)} />
	{/if}
</div>

<style>
	.beta-message-container {
		padding: 1rem 0;

		background-color: var(--ssb-green-1);
	}
	.beta-message {
		margin: 0;
		padding-left: 4rem;

		display: flex;
		align-items: center;
	}
</style>
