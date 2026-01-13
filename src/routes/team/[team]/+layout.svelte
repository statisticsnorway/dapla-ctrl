<script lang="ts">
	import { Alert } from '@nais/ds-svelte-community';
	import { browser } from '$app/environment';
	import type { LayoutProps } from './$types';
	import Menu from './Menu.svelte';
	import { createTeamContext } from './teamContext.svelte';
	import { page } from '$app/state';
	import EditTeamDisplayName from '$lib/ui/EditTeamDisplayName.svelte';

	let { data, children }: LayoutProps = $props();
	let { deletionInProgress, lastSuccessfulSync, UserInfo, viewerIsMember, displayName, teamSlug } =
		$derived(data);

	const section = $derived(browser ? page.data?.section : undefined);

	createTeamContext();

	const isAdmin = $derived(
		$UserInfo.data?.me.__typename === 'User' ? $UserInfo.data?.me.isAdmin : false
	);

	const viewerId = $derived(
		$UserInfo.data?.me.__typename === 'User' ? $UserInfo.data?.me.id : null
	);

	const isSectionManager = $derived(
		viewerId && section?.manager?.id ? viewerId === section.manager.id : false
	);

	const canEdit = $derived(isAdmin || isSectionManager);

	const isTeamOverviewPage = $derived(browser ? page.url.pathname === `/team/${teamSlug}` : false);
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
			<EditTeamDisplayName {displayName} {teamSlug} {canEdit} {isTeamOverviewPage} />
			<div>{@render children?.()}</div>
		</div>
	</div>
</div>

<style>
	.page {
		margin-top: var(--spacing-layout);
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
