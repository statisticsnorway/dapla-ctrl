<script lang="ts">
	import { page } from '$app/state';
	import { Alert, CopyButton } from '@nais/ds-svelte-community';
	import type { PageProps } from './$types';
	import TeamOverviewActivityLog from '$lib/domain/activity/team-overview/TeamOverviewActivityLog.svelte';
	let { data }: PageProps = $props();
	let { teamSlug, slug, section } = $derived(data);
</script>

{#if page.url.searchParams.has('deleted')}
	{@const msgParts = (page.url.searchParams.get('deleted') || '').split('/')}
	<Alert variant="success" size="small">
		Slettet {msgParts[0]}
		{msgParts[1]}.
	</Alert>
{/if}

<div class="main-layout">
	<div class="left-section">
		<div class="team-slug">
			<span class="slug-value">{slug || teamSlug}</span>
			<CopyButton
				copyText={slug || teamSlug}
				title={slug || teamSlug}
				iconPosition="right"
				size="xsmall"
			/>
		</div>
		<div class="spacer"></div>
		<div class="info-item">
			<div class="value">
				{#if section?.manager}
					{#if section.manager.email}
						<a href="/member/{section.manager.email}">{section.manager.name}</a>
					{:else}
						{section.manager.name}
					{/if}
				{:else}
					<span class="missing">Mangler seksjonsleder</span>
				{/if}
			</div>
			{#if section?.manager?.email}
				<div class="value">
					<a href="mailto:{section.manager.email}">{section.manager.email}</a>
				</div>
			{/if}
		</div>
		<div class="info-item">
			<div class="value">
				{#if section}
					{section.name} ({section.code})
				{:else}
					<span class="missing">Ikke spesifisert</span>
				{/if}
			</div>
		</div>
	</div>
	<div class="right-section">
		<TeamOverviewActivityLog {teamSlug} />
	</div>
</div>

<style>
	.main-layout {
		display: grid;
		grid-template-columns: 1fr auto;
		gap: var(--spacing-layout);
		align-items: start;
		margin-top: calc(-1 * var(--spacing-layout));
	}

	.left-section {
		display: flex;
		flex-direction: column;
		gap: var(--ax-space-4);
	}

	.team-slug {
		font-size: var(--ax-font-size-medium);
		color: var(--ax-text-subtle);
		display: flex;
		align-items: center;
		gap: var(--ax-space-8);
	}

	.slug-value {
		font-family: monospace;
	}

	.spacer {
		height: var(--ax-space-16);
	}

	.info-item {
		display: flex;
		flex-direction: column;
		gap: var(--ax-space-4);
	}

	.info-item .value {
		font-size: var(--ax-font-size-medium);
		color: var(--ax-text-default);
		display: flex;
		align-items: center;
		gap: var(--ax-space-8);
		flex-wrap: wrap;
	}

	.info-item .value a {
		color: var(--ax-text-action);
		text-decoration: none;
	}

	.info-item .value a:hover {
		text-decoration: underline;
	}

	.missing {
		color: var(--ax-text-subtle);
		font-style: italic;
	}

	.right-section {
		min-width: 300px;
		display: flex;
		flex-direction: column;
		gap: var(--ax-space-16);
	}
</style>
