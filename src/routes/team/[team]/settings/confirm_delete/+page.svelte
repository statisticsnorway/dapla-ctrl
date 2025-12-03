<script lang="ts">
	import { goto } from '$app/navigation';
	import {
		graphql,
		type ConfirmTeamDeletion$input,
		type ConfirmTeamDeletion$result,
		type QueryResult
	} from '$houdini';
	import GraphErrors from '$lib/ui/GraphErrors.svelte';
	import Time from '$lib/ui/Time.svelte';
	import { Alert, BodyLong, Button, Modal } from '@nais/ds-svelte-community';
	import { TrashIcon } from '@nais/ds-svelte-community/icons';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let { TeamDeleteKey, UserInfo } = $derived(data);

	let showConfirmDeleteTeam = $state(false);
	let deleteTeamLoading = $state(false);
	let deleteTeamResp: QueryResult<ConfirmTeamDeletion$result, ConfirmTeamDeletion$input> | null =
		$state(null);

	const deleteTeam = graphql(`
		mutation ConfirmTeamDeletion($key: String!, $team: Slug!) {
			confirmTeamDeletion(input: { key: $key, slug: $team }) {
				deletionStarted
			}
		}
	`);
</script>

<GraphErrors errors={$TeamDeleteKey.errors} />

{#if $TeamDeleteKey.data}
	{@const key = $TeamDeleteKey.data.team.deleteKey}
	{#if $UserInfo.data?.me.__typename == 'User' && $UserInfo.data.me.id == key.createdBy.id}
		<Alert variant="error">Du kan ikke bekrefte din egen slettingsforespørsel.</Alert>
	{:else if Date.now() - +key.expires > 0}
		<Alert variant="error">Slettingsnøkkelen har utløpt.</Alert>
	{:else}
		<BodyLong style="padding-bottom: 1rem;">
			Slettingen ble initiert av <strong>{key.createdBy.name}</strong> og utløper
			<strong><Time distance={true} time={key.expires}></Time></strong>. Å slette teamet vil
			permanent slette alle administrerte ressurser og alle ressurser innenfor dem. Alle
			applikasjoner, databaser og jobber eid av teamet vil bli uopprettelig slettet. Når du klikker
			på slett team er det ingen vei tilbake.
		</BodyLong>

		<Button
			onclick={() => {
				showConfirmDeleteTeam = !showConfirmDeleteTeam;
			}}
			variant="danger"
			icon={TrashIcon}
		>
			Slett team
		</Button>

		<Modal bind:open={showConfirmDeleteTeam} header="Bekreft team-sletting">
			<BodyLong>
				Bekreft at du har til hensikt å slette <strong>{key.team.slug}</strong> og alle ressurser relatert
				til det.
			</BodyLong>

			{#if deleteTeamResp?.errors}
				<GraphErrors errors={deleteTeamResp.errors} />
			{/if}

			{#snippet footer()}
				<Button
					type="submit"
					loading={deleteTeamLoading}
					onclick={async () => {
						deleteTeamLoading = true;
						deleteTeamResp = await deleteTeam.mutate({
							key: key.key,
							team: key.team.slug
						});
						goto('/team/' + key.team.slug, { replaceState: true });
					}}>Bekreft</Button
				>
				<Button
					variant="tertiary"
					disabled={deleteTeamLoading}
					type="reset"
					onclick={() => {
						showConfirmDeleteTeam = false;
					}}>Avbryt</Button
				>
			{/snippet}
		</Modal>
	{/if}
{/if}
