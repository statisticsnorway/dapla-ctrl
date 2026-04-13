<script lang="ts">
	import { afterNavigate, beforeNavigate } from '$app/navigation';
	import { page } from '$app/state';
	import { graphql } from '$houdini';
	import { isAuthenticated, isUnauthenticated } from '$lib/authentication';
	import '$lib/font-roboto.css';
	import '$lib/font-open-sans.css';
	import '$lib/font-roboto-condensed.css';
	import ProgressBar from '$lib/ui/ProgressBar.svelte';
	import { themeSwitch } from '$lib/stores/theme.svelte';
	import { Page, Theme } from '@nais/ds-svelte-community';
	import { onMount } from 'svelte';
	import '../styles/app.css';
	import '../styles/colors.css';
	import '../styles/ssb-colors.css';
	import '../styles/aksel-token-overrides.css';
	import type { LayoutProps } from './$houdini';
	import Login from './Login.svelte';
	import PageHeader from './PageHeader.svelte';

	let { data, children }: LayoutProps = $props();
	let { UserInfo, userAgent } = $derived(data);

	$effect(() => {
		themeSwitch.theme = data.theme;
	});

	let user = $derived(
		$UserInfo.data?.me as
			| {
					readonly name: string;
					readonly email: string;
					readonly isAdmin: boolean;
					readonly __typename: 'User';
			  }
			| undefined
	);

	const refreshCookie = graphql(`
		query RefreshCookie {
			me {
				__typename
			}
		}
	`);

	onMount(() => {
		setInterval(
			async () => {
				if (user?.__typename !== 'User') return;
				refreshCookie.fetch({ policy: 'NoCache' });
			},
			1000 * 60 * 10
		);
	});

	let loading = $state(false);

	beforeNavigate((navigation) => {
		if (navigation.from?.url.hostname === navigation.to?.url.hostname) {
			loading = true;
		}
	});

	afterNavigate(() => {
		loading = false;
	});

	const title = $derived.by(() => {
		const parts = [];
		if (page.data.meta.breadcrumbs && page.data.meta.breadcrumbs.length > 0) {
			parts.push(...page.data.meta.breadcrumbs.map((b) => b.label));
		}
		if (page.data.meta.title) {
			parts.unshift(page.data.meta.title);
		}
		return parts.join(' - ') + ' - Dapla Ctrl';
	});
</script>

<svelte:head>
	<title>
		{title}
	</title>
</svelte:head>

<Theme theme={themeSwitch.theme}>
	<Page contentBlockPadding="none">
		<div class="full-wrapper">
			{#if loading}
				<ProgressBar />
			{/if}

			{#if !$isAuthenticated || isUnauthenticated($UserInfo.errors)}
				<!-- logged out. We check both to support both  -->
				<Login />
			{:else}
				{#if user?.__typename === 'User'}
					<PageHeader {user} {userAgent} />
				{/if}

				{@render children?.()}
			{/if}
		</div>
	</Page>
</Theme>

<style>
	:global(.page) {
		margin-inline: 3rem;

		margin-top: var(--spacing-layout);
	}

	@media (max-width: 1464px) {
		:global(.page) {
			padding: 0 2rem;
		}
	}

	.full-wrapper {
		padding-bottom: 1rem;
	}
</style>
