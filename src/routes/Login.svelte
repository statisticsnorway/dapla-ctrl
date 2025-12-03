<script lang="ts">
	import { page } from '$app/stores';
	import { Alert, Button, Heading } from '@nais/ds-svelte-community';
	import Logo from '../Logo.svelte';

	const redirectPath = (url: URL) => {
		return encodeURIComponent(url.pathname + url.search + url.hash);
	};
</script>

<svelte:head>
	<title>Logg inn - Dapla Ctrl</title>
	<style>
		body {
			background: var(--ax-bg-default);
			background: linear-gradient(135deg, var(--ax-bg-default) 0%, var(--active-color) 100%);
		}
	</style>
</svelte:head>

<div class="wrapper">
	<div class="login">
		<Heading level="1" size="large" spacing>
			<Logo height=".8em" />
			Dapla Ctrl
		</Heading>
		{#if $page.url.searchParams?.get('error')}
			{@const error = $page.url.searchParams.get('error')}
			<Alert variant="error">
				{#if error == 'unknown-user'}
					Feil under innlogging: Ukjent bruker.<br />
					Vennligst kontakt systemadministratoren.
				{:else}
					<!-- "unable-to-create-session", "invalid-state", and "unauthenticated" are known. -->
					Feil under innlogging, vennligst prøv igjen.
				{/if}
			</Alert>
		{/if}

		<p>For å få tilgang til denne siden må du logge inn med din Google Workspace-konto.</p>

		<Button as="a" href="/oauth2/login?redirect_uri={redirectPath($page.url)}" variant="primary">
			Logg inn på Dapla Ctrl
		</Button>
	</div>
</div>

<style>
	.wrapper {
		display: flex;
		align-items: center;
		justify-content: center;
		height: calc(100vh - 1rem);
	}

	.login {
		text-align: center;
		max-width: 600px;
	}
</style>
