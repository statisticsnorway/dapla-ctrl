<script lang="ts">
	import { graphql } from '$houdini';
	import ExternalLink from '$lib/ui/ExternalLink.svelte';
	import GraphErrors from '$lib/ui/GraphErrors.svelte';
	import List from '$lib/ui/List.svelte';
	import ListItem from '$lib/ui/ListItem.svelte';
	import Pagination from '$lib/ui/Pagination.svelte';
	import { Alert, Button, Heading, Modal, TextField } from '@nais/ds-svelte-community';
	import { PlusIcon } from '@nais/ds-svelte-community/icons';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let { ArtifactRegistryGithubRepos, teamSlug, viewerIsMember, isManaged, UserInfo } =
		$derived(data);
	let isAdmin = $derived($UserInfo.data?.me.__typename === 'User' && $UserInfo.data.me.isAdmin);
	let canAdd = $derived(!isManaged && (viewerIsMember || isAdmin));
	let addModalOpen = $state(false);
	let repositoryName = $state('');
	let inputError = $state('');
	let mutationErrors: { message: string }[] | undefined = $state();
	let refreshErrors: { message: string }[] | undefined = $state();
	let successMessage = $state('');
	let submitting = $state(false);

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
</script>

<GraphErrors errors={$ArtifactRegistryGithubRepos.errors} />
<GraphErrors errors={refreshErrors} />
{#if successMessage}
	<Alert variant="success" size="small">{successMessage}</Alert>
{/if}
{#if $ArtifactRegistryGithubRepos.data}
	{@const repositories = $ArtifactRegistryGithubRepos.data.team.artifactRegistryAllowedGithubRepos}
	<List title={`GitHub-repositorier med tilgang (${repositories.pageInfo.totalCount})`}>
		{#snippet menu()}
			{#if canAdd}
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

<style>
	.actions {
		display: flex;
		gap: var(--ax-space-8);
		margin-top: var(--ax-space-16);
	}
</style>
