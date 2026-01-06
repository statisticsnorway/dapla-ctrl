<script lang="ts">
	import { themeSwitch } from '$lib/stores/theme.svelte';
	import {
		ActionMenu,
		ActionMenuCheckboxItem,
		ActionMenuDivider,
		ActionMenuItem,
		InternalHeaderButton,
		InternalHeaderTitle,
		InternalHeaderUserButton
	} from '@nais/ds-svelte-community/experimental';
	import { CogIcon } from '@nais/ds-svelte-community/icons';
	import SearchButton from '$lib/domain/search/SearchButton.svelte';

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

<header class="aksel-internalheader header">
	<InternalHeaderTitle as="a" href="/" style="border: none; padding: 0rem;">
		<div class="title">Dapla Ctrl</div>
	</InternalHeaderTitle>

	<div class="aksel-stack__spacer aksel-stack__spacer"></div>

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
		<ActionMenuCheckboxItem
			checked={themeSwitch.theme == 'dark'}
			onchange={(checked) => {
				if (!checked) {
					themeSwitch.setTheme('light');
				} else {
					themeSwitch.setTheme('dark');
				}
			}}
		>
			Mørkt tema
		</ActionMenuCheckboxItem>
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
</style>
