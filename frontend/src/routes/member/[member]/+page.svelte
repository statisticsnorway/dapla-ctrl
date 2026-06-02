<script lang="ts">
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let { UserOverview } = $derived(data);
</script>

{#if $UserOverview.data}
	<div class="user-info">
		<div class="info-section">
			<div class="value">
				<a href="mailto:{$UserOverview.data.user.email}">{$UserOverview.data.user.email}</a>
			</div>

			{#if $UserOverview.data.user.jobTitle}
				<div class="value">{$UserOverview.data.user.jobTitle}</div>
			{/if}

			{#if $UserOverview.data.user.section}
				<div class="value">
					{$UserOverview.data.user.section.name} ({$UserOverview.data.user.section.code})
				</div>
			{/if}
		</div>
	</div>
{/if}

<style>
	.user-info {
		margin-top: var(--ax-space-8);
		margin-bottom: var(--spacing-layout);
	}

	.info-section {
		display: flex;
		flex-direction: column;
		gap: var(--ax-space-8);
	}

	.value {
		font-size: var(--ax-font-size-medium);
		color: var(--ax-text-neutral);
	}

	.value a {
		color: var(--ax-text-neutral);
		text-decoration: none;
	}

	.value a:hover {
		text-decoration: underline;
	}
</style>
