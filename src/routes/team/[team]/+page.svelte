<script lang="ts">
	import { page } from '$app/state';
	import { Alert, CopyButton, Heading } from '@nais/ds-svelte-community';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let { activityLog, teamSlug, purpose } = $derived(data);
</script>

<div class="team-info">
	<div>
		{teamSlug}
		<CopyButton copyText={teamSlug} title={teamSlug} iconPosition="right" size="xsmall" />
	</div>
	<div>
		{purpose}
	</div>
</div>

{#if page.url.searchParams.has('deleted')}
	{@const msgParts = (page.url.searchParams.get('deleted') || '').split('/')}
	<Alert variant="success" size="small">
		Slettet {msgParts[0]}
		{msgParts[1]}.
	</Alert>
{/if}

<div class="wrapper">
	<div class="grid">
		<div class="card activity">
			<Heading size="small" level="2">Activity</Heading>
			{#if activityLog}
				<div class="raised">
					{#each activityLog.nodes as item (item.id)}
						<div>{item.__typename}</div>
					{/each}
				</div>
			{/if}
		</div>
	</div>
</div>

<style>
	.wrapper {
		display: grid;
		grid-template-columns: 1fr 300px;
		gap: var(--spacing-layout);
	}

	.team-info {
		display: grid;
		grid-template-columns: 1fr;
		margin-top: calc(-1 * var(--spacing-layout));
		margin-bottom: var(--spacing-layout);
	}
</style>
