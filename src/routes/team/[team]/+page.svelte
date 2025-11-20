<script lang="ts">
	import { page } from '$app/state';
	import ActivityLogItem from '$lib/components/ActivityLogItem.svelte';
	import { Alert, Heading } from '@nais/ds-svelte-community';
	import type { PageProps } from './$houdini';

	let { data }: PageProps = $props();
	let { TeamOverview, teamSlug, viewerIsMember } = $derived(data);
</script>

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
		display: flex;
		flex-direction: column;
		gap: var(--spacing-layout);
	}
	.raised {
		border-radius: 8px;
		display: flex;
		flex-direction: column;
		gap: 2px;

		> div {
			background-color: var(--ax-bg-default, --a-surface-default);
			padding: var(--ax-space-8, --a-spacing-2) var(--ax-space-20, --a-spacing-5);
		}

		> div:first-child {
			border-top-left-radius: 8px;
			border-top-right-radius: 8px;
			padding-top: var(--ax-space-12, --a-spacing-3);
		}

		> div:last-child {
			padding-bottom: var(--ax-space-12, --a-spacing-3);
			border-bottom-left-radius: 8px;
			border-bottom-right-radius: 8px;
		}
	}

	.card {
		background-color: var(--ax-bg-sunken, --a-surface-subtle);
		padding: var(--ax-space-16, --a-spacing-4) var(--ax-space-20, --a-spacing-5);
		border-radius: 12px;
		align-items: stretch;
	}

	.activity {
		grid-column: span 2;
		word-wrap: break-word;
		display: flex;
		flex-direction: column;
		gap: var(--ax-space-16, --a-spacing-4);
		min-height: 100%;

		> a {
			align-self: end;
		}
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
		gap: 1rem;
		grid-auto-flow: dense;
	}
	.grid:not(:first-child) {
		margin-top: 1rem;
	}
	.alerts-wrapper {
		display: flex;
		flex-direction: column;
		gap: var(--ax-space-8, --a-spacing-2);
	}
</style>
