<script lang="ts">
	import { page } from '$app/state';
	import { AlertState } from '$houdini';
	import GraphErrors from '$lib/ui/GraphErrors.svelte';
	import { Alert, BodyShort } from '@nais/ds-svelte-community';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let { TeamOverview, teamSlug, purpose } = $derived(data);
</script>

<div class="team-info">
	<div>
		<BodyShort>{purpose}</BodyShort>
	</div>
</div>

{#if page.url.searchParams.has('deleted')}
	{@const msgParts = (page.url.searchParams.get('deleted') || '').split('/')}
	<Alert variant="success" size="small">
		Successfully deleted {msgParts[0]}
		{msgParts[1]}.
	</Alert>
{/if}

<div class="wrapper">
	<div class="grid">
		<div class="card activity">
			<Heading size="small" level="2">Activity</Heading>
			{#if $TeamOverview.data}
				<div class="raised">
					{#each $TeamOverview.data.team.activityLog.nodes as item (item.id)}
						<div><ActivityLogItem {item} /></div>
					{/each}
				</div>
			{/if}
			{#if viewerIsMember}
				<a href="/team/{teamSlug}/activity-log" style:align-self="end" style:margin-top="auto"
					>View Activity Log</a
				>
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
		grid-template-columns: 1fr 300px;
		gap: var(--spacing-layout);
		margin-top: calc(-1 * var(--spacing-layout));
		margin-bottom: var(--spacing-layout);
	}
</style>
