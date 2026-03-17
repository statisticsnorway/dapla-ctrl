<script lang="ts">
	import { page } from '$app/state';
	import { Alert, Button, CopyButton, Heading, Tooltip } from '@nais/ds-svelte-community';
	import type { PageProps } from './$types';
	import TeamOverviewActivityLog from '$lib/domain/activity/team-overview/TeamOverviewActivityLog.svelte';
	import { PlusIcon, TrashIcon } from '@nais/ds-svelte-community/icons';
	import AddAccessManager from './AddAccessManager.svelte';
	import { graphql } from '$houdini';
	import Confirm from '$lib/ui/Confirm.svelte';
	let { data }: PageProps = $props();
	let { teamSlug, TeamOverview, UserInfo } = $derived(data);

	let team = $derived($TeamOverview.data?.team);

	let addAccessManagerOpen = $state(false);

	let removeAccessManagerOpen = $state(false);
	let removeAccessManager: { email: string; name: string } | null = $state(null);

	let canManageTeam = $derived.by(() => {
		let me = $UserInfo.data?.me;
		if (me?.__typename !== 'User') return false;
		return me.isAdmin || team?.section.manager?.email === me.email;
	});

	const refetch = () => {
		TeamOverview.fetch({
			policy: 'CacheAndNetwork'
		});
	};

	const removeTeamAccessManager = graphql(`
		mutation RemoveTeamAccessManager($input: RemoveTeamAccessManagerInput!) {
			removeTeamAccessManager(input: $input) {
				team {
					slug
				}
			}
		}
	`);
</script>

{#if page.url.searchParams.has('deleted')}
	{@const msgParts = (page.url.searchParams.get('deleted') || '').split('/')}
	<Alert variant="success" size="small">
		Slettet {msgParts[0]}
		{msgParts[1]}.
	</Alert>
{/if}

{#if team}
	{@const section = team.section}
	<div class="main-layout">
		<div class="left-section">
			<div class="team-slug">
				<span class="slug-value">{team.slug || teamSlug}</span>
				<CopyButton
					copyText={team.slug || teamSlug}
					title={team.slug || teamSlug}
					iconPosition="right"
					size="xsmall"
				/>
			</div>
			<div class="info-item">
				<div class="value">
					{section.name} ({section.code})
				</div>
			</div>
			<div class="spacer"></div>
			<div class="info-item">
				<Heading level="2" as="h2" size="xsmall">Teamansvarlig</Heading>
				<div class="value">
					{#if section.manager}
						<a href="/member/{section.manager.email}">{section.manager.name} ({section.code})</a>
					{:else}
						<span class="missing">Mangler seksjonsleder</span>
					{/if}
				</div>
			</div>

			<Heading level="2" as="h2" size="xsmall">
				<Tooltip content="Tilgangsansvarlige kan legge til og fjerne medlemmer fra teamet"
					>Tilgangsansvarlige</Tooltip
				>
				{#if canManageTeam}
					{@const canAddMore = team.accessManagers.length < 2}
					<Tooltip
						content={canAddMore
							? 'Legg til ny tilgangsansvarlig'
							: 'Teamet kan ha maks 2 tilgangsansvarlige i tillegg til den teamansvarlige'}
					>
						<Button
							disabled={!canAddMore}
							icon={PlusIcon}
							size="xsmall"
							onclick={() => {
								addAccessManagerOpen = !addAccessManagerOpen;
							}}
						/>
					</Tooltip>
				{/if}
			</Heading>
			<div>
				<div class="info-item">
					<div class="value">
						{#if section.manager}
							<a href="/member/{section.manager.email}">{section.manager.name} ({section.code})</a>
						{:else}
							<span class="missing">Mangler seksjonsleder</span>
						{/if}
					</div>
				</div>
				{#each team.accessManagers as am (am.user.id)}
					<div class="info-item">
						<div class="value">
							<a href="/member/{am.user.email}"
								>{am.user.name} ({am.user.section?.code ?? 'Mangler seksjon'})</a
							>
							{#if canManageTeam}<Button
									icon={TrashIcon}
									size="xsmall"
									variant="tertiary"
									onclick={() => {
										removeAccessManager = { email: am.user.email, name: am.user.name };
										removeAccessManagerOpen = !removeAccessManagerOpen;
									}}
								/>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		</div>
		<div class="right-section">
			<TeamOverviewActivityLog {teamSlug} />
		</div>
	</div>

	{#if addAccessManagerOpen}
		<AddAccessManager bind:open={addAccessManagerOpen} team={team.slug} oncreated={refetch} />
	{/if}
	{#if removeAccessManager && removeAccessManagerOpen}
		{@const userId = removeAccessManager.email}
		<Confirm
			bind:open={removeAccessManagerOpen}
			confirmText="Fjern"
			variant="danger"
			onconfirm={async () => {
				await removeTeamAccessManager.mutate({
					input: { teamSlug: team.slug, userEmail: userId }
				});
				refetch();
			}}
		>
			{#snippet header()}
				<Heading>Fjern tilgangsansvarlig</Heading>
			{/snippet}
			Er du sikker på at du vil fjerne <b>{removeAccessManager.name}</b> som tilgangsansvarlig?
		</Confirm>
	{/if}
{/if}

<style>
	.main-layout {
		display: grid;
		grid-template-columns: 1fr auto;
		gap: var(--spacing-layout);
		align-items: start;
		margin-top: calc(-1 * var(--spacing-layout));
	}

	.left-section {
		display: flex;
		flex-direction: column;
		gap: var(--ax-space-4);
	}

	.team-slug {
		font-size: var(--ax-font-size-medium);
		color: var(--ax-text-subtle);
		display: flex;
		align-items: center;
		gap: var(--ax-space-8);
	}

	.slug-value {
		font-family: monospace;
	}

	.spacer {
		height: var(--ax-space-16);
	}

	.info-item {
		display: flex;
		flex-direction: column;
		gap: var(--ax-space-4);
	}

	.info-item .value {
		font-size: var(--ax-font-size-medium);
		color: var(--ax-text-default);
		display: flex;
		align-items: center;
		gap: var(--ax-space-8);
		flex-wrap: wrap;
	}

	.info-item .value a {
		color: var(--ax-text-action);
		text-decoration: none;
	}

	.info-item .value a:hover {
		text-decoration: underline;
	}

	.missing {
		color: var(--ax-text-subtle);
		font-style: italic;
	}

	.right-section {
		min-width: 300px;
		display: flex;
		flex-direction: column;
		gap: var(--ax-space-16);
	}
</style>
