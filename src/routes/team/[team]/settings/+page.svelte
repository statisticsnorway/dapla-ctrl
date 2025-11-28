<script lang="ts">
	import { browser } from '$app/environment';
	import {
		type GetTeamDeleteKey$input,
		type GetTeamDeleteKey$result,
		graphql,
		type QueryResult
	} from '$houdini';
	import GraphErrors from '$lib/GraphErrors.svelte';
	import WarningIcon from '$lib/icons/WarningIcon.svelte';
	import {
		Alert,
		BodyLong,
		Button,
		CopyButton,
		Heading,
		Modal,
		TextField
	} from '@nais/ds-svelte-community';
	import {
		TrashIcon
	} from '@nais/ds-svelte-community/icons';
	import type { PageProps } from './$types';
	import EditText from './EditText.svelte';

	let { data }: PageProps = $props();
	let { TeamSettings, viewerIsOwner, teamSlug, viewerIsMember } = $derived(data);

	const updateTeam = graphql(`
		mutation UpdateTeam($input: UpdateTeamInput!) {
			updateTeam(input: $input) {
				team {
					purpose
				}
			}
		}
	`);

	const getTeamDeleteKey = graphql(`
		mutation GetTeamDeleteKey($input: RequestTeamDeletionInput!) {
			requestTeamDeletion(input: $input) {
				key {
					createdAt
					createdBy {
						email
					}
					expires
					key
					team {
						slug
					}
				}
			}
		}
	`);

	let deleteKeyLoading = $state(false);
	let deleteKeyResp: QueryResult<GetTeamDeleteKey$result, GetTeamDeleteKey$input> | null =
		$state(null);

	let teamSettings = $derived($TeamSettings.data?.team);

	let showDeleteTeam = $state(false);

	let descriptionErrors: { message: string }[] | undefined = $state();
</script>

<GraphErrors errors={$TeamSettings.errors} />

{#if teamSettings}
	<div class="wrapper">
		<div style="display: flex; flex-direction: column; gap: var(--spacing-layout)">
			<div>
				<Heading level="2">Description</Heading>
				<EditText
					text={teamSettings.purpose}
					on:save={async (e) => {
						descriptionErrors = undefined;
						const data = await updateTeam.mutate({
							input: {
								slug: teamSlug,
								purpose: e.detail
							}
						});

						if (data.errors) {
							descriptionErrors = data.errors;
						}
					}}
					isMember={viewerIsMember}
				/>

				<GraphErrors errors={descriptionErrors} size="small" />
			</div>

			{#if viewerIsOwner}
				<div>
					<Heading level="2"><WarningIcon class="heading-aligned-icon" /> Danger Zone</Heading>
					<div class="danger-zone">
						<BodyLong spacing>
							Deleting the team will permanently delete all managed resources and all resources
							within them. All applications, databases and jobs owned by the team will be
							irreversibly deleted.
						</BodyLong>
						<BodyLong spacing>
							When you request deletion a delete key will be generated for this team. It is valid
							for 1 hour. Another team-owner will have to confirm the deletion by using a generated
							link before the team is irreversibly deleted.
						</BodyLong>

						<Button
							variant="danger"
							onclick={() => {
								showDeleteTeam = !showDeleteTeam;
								//deleteKeyResp = null;
							}}
							icon={TrashIcon}
						>
							Request team deletion</Button
						>
					</div>
				</div>
			{/if}
		</div>
	</div>
{/if}
{#if browser}
	<Modal bind:open={showDeleteTeam}>
		{#snippet header()}
			<Heading level="1" size="medium">Request Team Deletion</Heading>
		{/snippet}

		{#if !deleteKeyResp?.data}
			<BodyLong>
				Confirm that you intend to delete <strong>{teamSlug}</strong> and all resources related to it.
			</BodyLong>
		{/if}

		{#if deleteKeyResp?.errors}
			<GraphErrors errors={deleteKeyResp.errors}></GraphErrors>
		{:else if deleteKeyResp?.data}
			{@const key =
				window.location + '/confirm_delete?key=' + deleteKeyResp.data.requestTeamDeletion.key?.key}
			<Alert variant="info">
				Deletion of <strong>{teamSlug}</strong> has been requested. To finalize the deletion send
				this link to another team owner and let them confirm the deletion.

				<div class="deletewrapper">
					<div>
						<TextField
							label="Sharable url"
							hideLabel={true}
							readonly={true}
							size="small"
							value={key}
						></TextField>
					</div>
					<CopyButton
						text="Copy URL"
						activeText="URL copied"
						variant="action"
						copyText={key}
						size="small"
					/>
				</div>
			</Alert>
		{/if}

		{#snippet footer()}
			{#if !deleteKeyResp?.data}
				<Button
					type="submit"
					loading={deleteKeyLoading}
					onclick={async () => {
						deleteKeyLoading = true;
						deleteKeyResp = await getTeamDeleteKey.mutate({
							input: { slug: teamSlug }
						});
						deleteKeyLoading = false;
					}}>Confirm</Button
				>
				<Button
					variant="tertiary"
					disabled={deleteKeyLoading}
					type="reset"
					onclick={() => {
						showDeleteTeam = !showDeleteTeam;
					}}>Cancel</Button
				>
			{:else}
				<Button
					onclick={() => {
						showDeleteTeam = !showDeleteTeam;
					}}>Close</Button
				>
			{/if}
		{/snippet}
	</Modal>
{/if}

<style>
	.wrapper {
		display: grid;
		grid-template-columns: 1fr 320px;
		gap: var(--spacing-layout);
	}
	.danger-zone {
		padding: var(--ax-space-16);
		border-radius: 8px;
		border: 1px solid var(--ax-border-danger);
	}

	.deployKey {
		font-family: monospace;
		padding-bottom: 1rem;
	}

	.buttons {
		display: flex;
		flex-direction: row;
		gap: 1rem;
	}
	.button {
		width: 130px;
	}

	.channel {
		display: flex;
		flex-direction: row;
		gap: 0.5rem;
	}

	.deletewrapper {
		display: flex;
		gap: 0.2rem;
	}

	.deletewrapper div {
		flex-grow: 1;
	}
</style>
