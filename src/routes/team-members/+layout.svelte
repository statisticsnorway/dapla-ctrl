<script lang="ts">
	import { page } from '$app/state';
	import { BodyShort, Heading } from '@nais/ds-svelte-community';
	import Tab from '$lib/ui/Tab.svelte';
	import Tabs from '$lib/ui/Tabs.svelte';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();
	let { MyTeamMembers, AllTeamMembers } = $derived(data);

	let myTeamMembersCount = $derived(
		$MyTeamMembers.data?.me.__typename === 'User'
			? $MyTeamMembers.data.me.teamMembers.pageInfo.totalCount
			: 0
	);

	let allTeamMembersCount = $derived($AllTeamMembers.data?.users?.pageInfo?.totalCount ?? 0);

	const isAllTeamMembersPage = $derived(page.url.pathname === '/team-members/all');

	const description = $derived(
		isAllTeamMembersPage
			? 'Oversikt over alle brukere i SSB som medlem av minst ett team.'
			: 'Oversikt over alle brukere som medlem av samme team som deg.'
	);
</script>

<svelte:head>
	<title>Medlemmer - Dapla Ctrl</title>
</svelte:head>

<div class="page">
	<div class="content-wrapper">
		<div class="header">
			<div>
				<Heading level="1" size="xlarge">Medlemmer</Heading>
				<div class="description">
					<BodyShort textColor="subtle" size="medium">{description}</BodyShort>
				</div>
			</div>
		</div>
		<div class="container" data-sveltekit-preload-data="hover">
			<div>
				<Tabs>
					<Tab
						data-sveltekit-noscroll
						href="/team-members"
						active={page.url.pathname === '/team-members' || page.url.pathname === '/team-members/'}
						title="Mine teammedlemmer ({myTeamMembersCount})"
					/>
					<Tab
						data-sveltekit-noscroll
						href="/team-members/all"
						active={page.url.pathname === '/team-members/all'}
						title="Alle teammedlemmer ({allTeamMembersCount})"
					/>
				</Tabs>

				{@render children()}
			</div>
		</div>
	</div>
</div>

<style>
	.page {
		margin-inline: var(--margin-default);
	}

	.content-wrapper {
		background: var(--ax-bg-default);
		position: relative;
	}

	.header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--ax-space-16);
	}

	.description {
		margin-top: var(--ax-space-4);
	}

	.container {
		margin-top: var(--spacing-layout);
		display: flex;
		flex-direction: column;
		gap: var(--spacing-layout);
	}
</style>
