<script lang="ts">
	import { graphql } from '$houdini';
	import ExternalLink from '$lib/ui/ExternalLink.svelte';
	import GraphErrors from '$lib/ui/GraphErrors.svelte';
	import List from '$lib/ui/List.svelte';
	import ListItem from '$lib/ui/ListItem.svelte';
	import Pagination from '$lib/ui/Pagination.svelte';
	import { Alert, BodyShort, Button, Heading, Modal, TextField } from '@nais/ds-svelte-community';
	import { PlusIcon, TrashIcon } from '@nais/ds-svelte-community/icons';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let { ArtifactRegistryGithubRepos, teamSlug, viewerIsMember, isManaged, UserInfo } =
		$derived(data);
	let isAdmin = $derived($UserInfo.data?.me.__typename === 'User' && $UserInfo.data.me.isAdmin);
	let canManageAccess = $derived(!isManaged && (viewerIsMember || isAdmin));
	let addModalOpen = $state(false);
	let removeModalOpen = $state(false);
	let removeRepositoryName = $state('');
	let repositoryName = $state('');
	let inputError = $state('');
	let mutationErrors: { message: string }[] | undefined = $state();
	let removeErrors: { message: string }[] | undefined = $state();
	let refreshErrors: { message: string }[] | undefined = $state();
	let successMessage = $state('');
	let submitting = $state(false);
	let removing = $state(false);

	const grantAccess = graphql(`
		mutation GrantGithubRepoAccess($input: GrantGithubRepoAccessToTeamArtifactRegistryInput!) {
			grantGithubRepoAccessToTeamArtifactRegistry(input: $input) {
				repository {
					id
					name
				}
			}
		}
	`);
	const revokeAccess = graphql(`
		mutation RevokeGithubRepoAccess($input: RevokeGithubRepoAccessFromTeamArtifactRegistryInput!) {
			revokeGithubRepoAccessFromTeamArtifactRegistry(input: $input) {
				success
			}
		}
	`);
	const addRepository = async () => {
		if (submitting) return;
		const name = repositoryName.trim();
		if (!name || name.includes('/')) {
			inputError = 'Skriv kun repository-navnet, uten statisticsnorway/.';
			return;
		}
		submitting = true;
		mutationErrors = undefined;
		refreshErrors = undefined;
		successMessage = '';
		inputError = '';
		let added = false;
		try {
			const result = await grantAccess.mutate({ input: { teamSlug, repositoryName: name } });
			if (result.errors?.length) {
				mutationErrors = result.errors.map(({ message }) => ({ message }));
				return;
			}
			if (!result.data?.grantGithubRepoAccessToTeamArtifactRegistry.repository) {
				mutationErrors = [{ message: 'Kunne ikke legge til repository.' }];
				return;
			}
			addModalOpen = false;
			repositoryName = '';
			successMessage = `statisticsnorway/${name} er lagt til. Tilgangen kan ta litt tid å aktivere.`;
			added = true;
		} catch {
			mutationErrors = [{ message: 'Kunne ikke legge til repository. Prøv igjen.' }];
		} finally {
			submitting = false;
		}
		if (added) {
			try {
				await ArtifactRegistryGithubRepos.fetch({ policy: 'NetworkOnly' });
			} catch {
				refreshErrors = [{ message: 'Tilgangen ble lagt til, men listen kunne ikke oppdateres.' }];
			}
		}
	};

	const removeRepository = async () => {
		if (removing || !removeRepositoryName) return;

		removing = true;
		removeErrors = undefined;
		refreshErrors = undefined;
		successMessage = '';
		const name = removeRepositoryName;
		let removed = false;
		try {
			const result = await revokeAccess.mutate({ input: { teamSlug, repositoryName: name } });
			if (result.errors?.length) {
				removeErrors = result.errors.map(({ message }) => ({ message }));
				return;
			}
			if (!result.data?.revokeGithubRepoAccessFromTeamArtifactRegistry.success) {
				removeErrors = [{ message: 'Kunne ikke fjerne tilgang.' }];
				return;
			}
			removeModalOpen = false;
			removeRepositoryName = '';
			successMessage = `Tilgangen for statisticsnorway/${name} er fjernet`;
			removed = true;
		} catch {
			removeErrors = [{ message: 'Kunne ikke fjerne tilgang. Prøv igjen.' }];
		} finally {
			removing = false;
		}
		if (removed) {
			try {
				await ArtifactRegistryGithubRepos.fetch({ policy: 'NetworkOnly' });
			} catch {
				refreshErrors = [{ message: 'Tilgangen ble fjernet, men listen kunne ikke oppdateres.' }];
			}
		}
	};
</script>

<div class="description">
	<BodyShort textColor="subtle" size="medium">
		Oversikt over Github-repositorier som har tilgang til å skrive til teamets Artifact Registry.
	</BodyShort>
</div>

<GraphErrors errors={$ArtifactRegistryGithubRepos.errors} />
<GraphErrors errors={refreshErrors} />
{#if successMessage}
	<Alert variant="success" size="small">{successMessage}</Alert>
{/if}
{#if $ArtifactRegistryGithubRepos.data}
	{@const repositories = $ArtifactRegistryGithubRepos.data.team.artifactRegistryAllowedGithubRepos}
	<List title={`GitHub-repositorier med tilgang (${repositories.pageInfo.totalCount})`}>
		{#snippet menu()}
			{#if canManageAccess}
				<Button
					size="small"
					variant="secondary"
					icon={PlusIcon}
					onclick={() => {
						mutationErrors = undefined;
						inputError = '';
						addModalOpen = true;
					}}
				>
					Legg til repository
				</Button>
			{/if}
		{/snippet}
		{#if repositories.edges.length === 0}
			<ListItem>Ingen GitHub-repositorier har tilgang ennå.</ListItem>
		{:else}
			{#each repositories.edges as { node: repository } (repository.id)}
				<ListItem>
					<ExternalLink href={`https://github.com/statisticsnorway/${repository.name}`}>
						statisticsnorway/{repository.name}
					</ExternalLink>
					{#if canManageAccess}
						<Button
							size="small"
							variant="tertiary"
							icon={TrashIcon}
							aria-label={`Fjern tilgang for statisticsNorway/${repository.name}`}
							onclick={() => {
								removeErrors = undefined;
								removeRepositoryName = repository.name;
								removeModalOpen = true;
							}}
						/>
					{/if}
				</ListItem>
			{/each}
		{/if}
	</List>
	<Pagination
		page={repositories.pageInfo}
		loaders={{
			loadPreviousPage: () => ArtifactRegistryGithubRepos.loadPreviousPage(),
			loadNextPage: () => ArtifactRegistryGithubRepos.loadNextPage()
		}}
	/>
{/if}
<Modal bind:open={addModalOpen}>
	{#snippet header()}
		<Heading level="2">Legg til GitHub-repository</Heading>
	{/snippet}
	<p>Gi et GitHub-repository tilgang til å skrive til teamets Artifact Registry.</p>
	<GraphErrors errors={mutationErrors} />
	<form
		onsubmit={(event: SubmitEvent) => {
			event.preventDefault();
			void addRepository();
		}}
	>
		<TextField
			id="artifact-registry-repository-name"
			type="text"
			bind:value={repositoryName}
			error={inputError || undefined}
			oninput={() => (inputError = '')}
		>
			{#snippet label()}Repository-navn{/snippet}
			{#snippet description()}Skriv kun navnet, for eksempel dapla-ctrl (uten statisticsnorway/).{/snippet}
		</TextField>
		<div class="actions">
			<Button type="submit" size="small" icon={PlusIcon} disabled={submitting}>Legg til</Button>
			<Button type="button" size="small" variant="tertiary" onclick={() => (addModalOpen = false)}>
				Avbryt
			</Button>
		</div>
	</form>
</Modal>

<Modal bind:open={removeModalOpen}>
	{#snippet header()}
		<Heading level="2">Fjern tilgang til Artifact Registry</Heading>
	{/snippet}
	<p>
		Er du sikker på at du vil fjerne skrivetilgang for
		<strong>statisticsnorway/{removeRepositoryName}</strong>? Dette sletter ikke
		GitHub-repositoriet.
	</p>
	<GraphErrors errors={removeErrors} />
	<div class="actions">
		<Button
			type="button"
			size="small"
			variant="danger"
			disabled={removing}
			onclick={() => void removeRepository()}
		>
			Fjern tilgang
		</Button>
		<Button
			type="button"
			size="small"
			variant="tertiary"
			disabled={removing}
			onclick={() => (removeModalOpen = false)}>Avbryt</Button
		>
	</div>
</Modal>

<style>
	.description {
		margin-top: calc(-1 * var(--spacing-layout));
		margin-bottom: var(--ax-space-16);
	}

	.actions {
		display: flex;
		gap: var(--ax-space-8);
		margin-top: var(--ax-space-16);
	}
</style>
