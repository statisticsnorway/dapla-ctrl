<script lang="ts">
	// import { page } from '$app/stores';
	import PageHeader from '$lib/ui/PageHeader.svelte';
	import { Alert } from '@nais/ds-svelte-community';
	import type { LayoutProps } from './$types';
	import Menu from './Menu.svelte';
	import { createTeamContext } from './teamContext.svelte';

	let { data, children }: LayoutProps = $props();
	let { deletionInProgress, lastSuccessfulSync, UserInfo, viewerIsMember } = $derived(data);

	createTeamContext();

	const isAdmin = $derived(
		$UserInfo.data?.me.__typename === 'User' ? $UserInfo.data?.me.isAdmin : false
	);
</script>

<div class="page">
	{#if deletionInProgress}
		<Alert variant="warning" style="margin-bottom: 1rem;"
			>Teamet og tilhørende ressurser er under sletting.</Alert
		>
	{/if}
	{#if !lastSuccessfulSync && !deletionInProgress}
		<Alert variant="info" style="margin-bottom: 1rem;" contentMaxWidth={false}
			>Teamet og tilhørende ressurser blir opprettet. Det tar vanligvis opptil 15 minutter.</Alert
		>
	{/if}

	<div class="main">
		<Menu member={viewerIsMember} {isAdmin} />
		<div class="container">
			<PageHeader />
			<div>{@render children?.()}</div>
		</div>
	</div>
</div>

<style>
	.page {
		margin-top: var(--spacing-layout);
		width: 100%;
	}

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
