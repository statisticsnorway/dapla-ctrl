<script lang="ts">
	import { page } from '$app/state';
	import { Alert, Button, Heading } from '@nais/ds-svelte-community';
	import { redirect } from '@sveltejs/kit';

	const redirectPath = (url: URL) => {
		return encodeURIComponent(url.pathname + url.search + url.hash);
	};

	const oauth2LoginPath = '/oauth2/login?redirect_uri=' + redirectPath(page.url);
	const errorParam = page.url.searchParams?.get('error');
	if (!errorParam) {
		redirect(302, oauth2LoginPath);
	}
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
		<Heading level="1" size="large" spacing>Dapla Ctrl</Heading>
		{#if errorParam}
			{@const error = page.url.searchParams.get('error')}
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

		<p>For å få tilgang til denne siden må du logge inn med din Entra ID-konto.</p>

		<Button as="a" href={oauth2LoginPath} variant="primary">Logg inn på Dapla Ctrl</Button>
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
