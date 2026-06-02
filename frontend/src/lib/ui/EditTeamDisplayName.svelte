<script lang="ts">
	import { graphql } from '$houdini';
	import PageHeader from '$lib/ui/PageHeader.svelte';
	import { Alert, Button, TextField } from '@nais/ds-svelte-community';
	import { PencilIcon } from '@nais/ds-svelte-community/icons';
	import { invalidateAll } from '$app/navigation';

	interface Props {
		displayName: string;
		teamSlug: string;
		canEdit: boolean;
		isTeamOverviewPage: boolean;
	}

	let { displayName, teamSlug, canEdit, isTeamOverviewPage }: Props = $props();

	const updateTeamDisplayName = graphql(`
		mutation UpdateTeamDisplayName($input: UpdateTeamInput!) {
			updateTeam(input: $input) {
				team {
					id
					slug
					displayName
				}
			}
		}
	`);

	let isEditing = $state(false);
	let newDisplayName = $state('');
	let updateError: string | null = $state(null);

	$effect(() => {
		if (!isEditing) {
			newDisplayName = displayName || '';
		}
	});

	$effect(() => {
		if (isEditing && !isTeamOverviewPage) {
			isEditing = false;
			updateError = null;
		}
	});

	const startEdit = () => {
		isEditing = true;
		newDisplayName = displayName || '';
		updateError = null;
	};

	const cancelEdit = () => {
		isEditing = false;
		newDisplayName = displayName || '';
		updateError = null;
	};

	const saveDisplayName = async () => {
		if (!newDisplayName.trim() || newDisplayName === displayName) {
			cancelEdit();
			return;
		}

		updateError = null;

		try {
			const result = await updateTeamDisplayName.mutate({
				input: {
					slug: teamSlug,
					displayName: newDisplayName.trim()
				}
			});

			if (result?.data?.updateTeam?.team) {
				isEditing = false;
				updateError = null;
				await invalidateAll();
			} else if (result?.errors) {
				const errorMessage = result.errors[0]?.message || 'Kunne ikke oppdatere team';
				updateError = errorMessage.includes('not authorized')
					? 'Du har ikke tilgang til å oppdatere dette teamet.'
					: errorMessage;
			}
		} catch (error) {
			updateError = 'En uventet feil oppstod ved oppdatering av teamet.';
			console.error('Error updating team:', error);
		}
	};
</script>

{#if isEditing && isTeamOverviewPage}
	<div class="edit-header">
		{#if updateError}
			<Alert variant="error" size="small" style="margin-bottom: var(--ax-space-8);">
				{updateError}
			</Alert>
		{/if}
		<div class="edit-wrapper">
			<div class="edit-textfield">
				<TextField
					label=""
					hideLabel
					size="medium"
					bind:value={newDisplayName}
					onkeydown={(e) => {
						if (e.key === 'Enter') {
							saveDisplayName();
						} else if (e.key === 'Escape') {
							cancelEdit();
						}
					}}
				/>
			</div>
			<Button onclick={saveDisplayName} size="xsmall" disabled={!newDisplayName.trim()}>
				Lagre
			</Button>
			<Button onclick={cancelEdit} size="xsmall" variant="secondary-neutral">Avbryt</Button>
		</div>
	</div>
{:else}
	{#snippet editActions()}
		{#if canEdit && isTeamOverviewPage && !isEditing}
			<Button
				onclick={startEdit}
				size="xsmall"
				variant="tertiary"
				icon={PencilIcon}
				aria-label="Endre visningsnavn"
			/>
		{/if}
	{/snippet}
	<PageHeader actions={editActions} />
{/if}

<style>
	.edit-wrapper {
		display: flex;
		align-items: stretch;
		gap: var(--ax-space-8);
	}

	.edit-wrapper :global(button) {
		align-self: center;
		min-height: calc(2rem * 1.2 + var(--ax-space-4) * 2);
		padding-top: var(--ax-space-4);
		padding-bottom: var(--ax-space-4);
	}

	.edit-textfield {
		flex: 0 1 auto;
		min-width: 300px;
		max-width: 600px;
	}

	.edit-textfield :global(.aksel-text-field) {
		width: 100%;
	}

	.edit-textfield :global(.aksel-text-field__input) {
		font-size: 2rem;
		font-weight: 600;
		line-height: 1.2;
		padding: var(--ax-space-4) var(--ax-space-8);
		width: 100%;
	}

	.edit-header {
		margin-bottom: var(--spacing-layout);
	}
</style>
