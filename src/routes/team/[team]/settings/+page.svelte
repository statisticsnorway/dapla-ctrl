<script lang="ts">
	import { browser } from '$app/environment';
	import {
		type GetTeamDeleteKey$input,
		type GetTeamDeleteKey$result,
		graphql,
		type QueryResult
	} from '$houdini';
	import GraphErrors from '$lib/ui/GraphErrors.svelte';
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
	import { TrashIcon } from '@nais/ds-svelte-community/icons';
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
				<Heading level="2">Beskrivelse</Heading>
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
					<Heading level="2"><WarningIcon class="heading-aligned-icon" /> Faresone</Heading>
					<div class="danger-zone">
						<BodyLong spacing>
							Å slette teamet vil permanent slette alle administrerte ressurser og alle ressurser
							innenfor dem. Alle applikasjoner, databaser og jobber eid av teamet vil bli
							uopprettelig slettet.
						</BodyLong>
						<BodyLong spacing>
							Når du ber om sletting, vil en slettingsnøkkel bli generert for dette teamet. Den er
							gyldig i 1 time. En annen team-eier må bekrefte slettingen ved å bruke en generert
							lenke før teamet blir uopprettelig slettet.
						</BodyLong>

						<Button
							variant="danger"
							onclick={() => {
								showDeleteTeam = !showDeleteTeam;
								//deleteKeyResp = null;
							}}
							icon={TrashIcon}
						>
							Be om team-sletting
						</Button>
					</div>
				</div>
			{/if}
		</div>
	</div>
{/if}
{#if browser}
	<Modal bind:open={showDeleteTeam}>
		{#snippet header()}
			<Heading level="1" size="medium">Be om team-sletting</Heading>
		{/snippet}

		{#if !deleteKeyResp?.data}
			<BodyLong>
				Bekreft at du har til hensikt å slette <strong>{teamSlug}</strong> og alle ressurser relatert
				til det.
			</BodyLong>
		{/if}

		{#if deleteKeyResp?.errors}
			<GraphErrors errors={deleteKeyResp.errors}></GraphErrors>
		{:else if deleteKeyResp?.data}
			{@const key =
				window.location + '/confirm_delete?key=' + deleteKeyResp.data.requestTeamDeletion.key?.key}
			<Alert variant="info">
				Sletting av <strong>{teamSlug}</strong> har blitt forespurt. For å fullføre slettingen, send
				denne lenken til en annen team-eier og la dem bekrefte slettingen.

				<div class="deletewrapper">
					<div>
						<TextField label="Delbar URL" hideLabel={true} readonly={true} size="small" value={key}
						></TextField>
					</div>
					<CopyButton
						text="Kopier URL"
						activeText="URL kopiert"
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
					}}>Bekreft</Button
				>
				<Button
					variant="tertiary"
					disabled={deleteKeyLoading}
					type="reset"
					onclick={() => {
						showDeleteTeam = !showDeleteTeam;
					}}>Avbryt</Button
				>
			{:else}
				<Button
					onclick={() => {
						showDeleteTeam = !showDeleteTeam;
					}}>Lukk</Button
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

	.deletewrapper {
		display: flex;
		gap: 0.2rem;
	}

	.deletewrapper div {
		flex-grow: 1;
	}
</style>
