<script lang="ts">
	// import { page } from '$app/stores';
	import PageHeader from '$lib/ui/PageHeader.svelte';
	import type { LayoutProps } from './$types';
	import Menu from './Menu.svelte';

	let { children }: LayoutProps = $props();
</script>

<div class="page">
	<div class="main">
		<Menu />
		<div class="container">
			<PageHeader />
			<div>{@render children?.()}</div>
		</div>
	</div>
</div>

<style>
	.main {
		gap: var(--spacing-layout);
		display: grid;
		grid-template-columns: 202px 1fr;
	}

	.container {
		flex-grow: 1;
		display: flex;
		flex-direction: column;
		gap: var(--spacing-layout);
	}
</style>
