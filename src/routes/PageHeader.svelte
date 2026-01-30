<script lang="ts">
	import SearchButton from '$lib/domain/search/SearchButton.svelte';
	import { themeSwitch } from '$lib/stores/theme.svelte';
	import { Button } from '@nais/ds-svelte-community';
	import {
		ActionMenu,
		ActionMenuDivider,
		ActionMenuItem,
		InternalHeaderButton,
		InternalHeaderTitle,
		InternalHeaderUserButton
	} from '@nais/ds-svelte-community/experimental';
	import { CogIcon, MoonIcon, SunIcon } from '@nais/ds-svelte-community/icons';
	import BetaBanner from './BetaBanner.svelte';

	interface Props {
		user:
			| {
					readonly name: string;
					readonly isAdmin: boolean;
			  }
			| undefined;
		userAgent: string;
	}

	let { user, userAgent }: Props = $props();
</script>

<BetaBanner />

<header class="aksel-internalheader header">
	<InternalHeaderTitle as="a" href="/" style="border: none; padding: 0rem; margin-right: 3rem;">
		<div class="title">Dapla Ctrl</div>
	</InternalHeaderTitle>
	<InternalHeaderButton as="a" href="/" style="font-size: var(--ax-font-size-medium);">
		Team
	</InternalHeaderButton>
	<InternalHeaderButton as="a" href="/team-members" style="font-size: var(--ax-font-size-medium);">
		Medlemmer
	</InternalHeaderButton>

	<InternalHeaderButton as="a" href="/shared-data">Datadeling</InternalHeaderButton>

	<div class="aksel-stack__spacer aksel-stack__spacer"></div>

	<Button
		style="background-color: inherit; color: inherit;"
		onclick={() => {
			if (themeSwitch.theme == 'dark') {
				themeSwitch.setTheme('light');
			} else {
				themeSwitch.setTheme('dark');
			}
		}}
	>
		<span class="switch-theme-icon">
			{#if themeSwitch.theme == 'dark'}
				<SunIcon />
			{:else}
				<MoonIcon />
			{/if}
		</span>
	</Button>

	<SearchButton {userAgent} />
	<InternalHeaderButton
		as="a"
		href="https://manual.dapla.ssb.no/statistikkere/dapla-ctrl"
		style="font-size: var(--ax-font-size-medium);"
	>
		Dokumentasjon
	</InternalHeaderButton>
	<ActionMenu>
		{#snippet trigger(props)}
			<InternalHeaderUserButton name={user ? user.name : 'unauthorized'} {...props} />
		{/snippet}

		{#if user?.isAdmin}
			<ActionMenuItem>
				<a href="/admin" class="action-menu-link" style="text-decoration: none;"><CogIcon />Admin</a
				></ActionMenuItem
			>
			<ActionMenuDivider />
		{/if}
	</ActionMenu>
</header>

<style>
	.header {
		height: 80px;
		padding-left: 4rem;
	}
	.title {
		color: var(--dapla-ctrl-logo);
		text-decoration: none;
		display: flex;
		align-items: center;
		gap: 0.25rem;
		font-family: 'Roboto Condensed';
		font-size: 2rem;
		font-weight: 700;
	}
	.action-menu-link {
		color: var(--ax-text-neutral);
		text-decoration: none;
	}
	.switch-theme-icon {
		display: flex;
		align-items: center;
		padding-top: 3px;
	}
</style>
