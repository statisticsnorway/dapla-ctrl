<script lang="ts">
	import { themeSwitch } from '$lib/stores/theme.svelte';
	import {
		ActionMenu,
		ActionMenuCheckboxItem,
		ActionMenuDivider,
		ActionMenuItem,
		InternalHeader,
		InternalHeaderButton,
		InternalHeaderTitle,
		InternalHeaderUserButton
	} from '@nais/ds-svelte-community/experimental';
	import { CogIcon, LeaveIcon } from '@nais/ds-svelte-community/icons';
	import Logo from '../Logo.svelte';
	import SearchButton from '$lib/domain/search/SearchButton.svelte';

	interface Props {
		user:
			| {
					readonly name: string;
					readonly isAdmin: boolean;
			  }
			| undefined;
	}

	let { user }: Props = $props();
</script>

<InternalHeader allowLightMode={true}>
	<InternalHeaderTitle as="a" href="/">
		<div class="logo">
			<Logo height="2.0rem" />
			<span>Ctrl</span>
		</div>
	</InternalHeaderTitle>

	<div class="aksel-stack__spacer aksel-stack__spacer"></div>

	<SearchButton />
	<InternalHeaderButton as="a" href="https://manual.dapla.ssb.no/statistikkere/dapla-ctrl"
		>Dokumentasjon</InternalHeaderButton
	>
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
		<ActionMenuItem>
			<a href="/oauth2/logout" class="action-menu-link" style="text-decoration: none;">
				<LeaveIcon />
				Logout
			</a>
		</ActionMenuItem>
	</ActionMenu>
</InternalHeader>

<style>
	.logo {
		color: var(--ax-text-neutral);
		text-decoration: none;
		display: flex;
		align-items: center;
		gap: 0.25rem;
		font-size: 1.5rem;
		font-weight: 700;
	}
	.action-menu-link {
		color: var(--ax-text-neutral);
		text-decoration: none;
	}
</style>
