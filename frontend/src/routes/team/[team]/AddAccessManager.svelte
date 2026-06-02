<script lang="ts">
	import { graphql, type AddTeamAccessManagerInput } from '$houdini';
	import { Alert, Button, Heading, Modal, TextField } from '@nais/ds-svelte-community';
	import { PlusIcon } from '@nais/ds-svelte-community/icons';
	import { onMount } from 'svelte';

	interface Props {
		open: boolean;
		team: string;
		oncreated?: () => void;
	}

	let { open = $bindable(), team, oncreated }: Props = $props();

	const store = graphql(`
		query AddTeamAccessManagerQuery($team: Slug!) {
			users(first: 10000) {
				nodes {
					id
					email
					name
				}
			}
			team(slug: $team) {
				section {
					manager {
						id
						email
					}
				}
				accessManagers {
					user {
						id
						email
					}
				}
			}
		}
	`);

	onMount(() => {
		store.fetch({
			variables: {
				team: team
			}
		});
	});

	const create = graphql(`
		mutation AddAccessManagerMutation($input: AddTeamAccessManagerInput!) {
			addTeamAccessManager(input: $input) {
				user {
					name
				}
			}
		}
	`);

	let users = $derived.by(() => {
		const allUsers = $store.data?.users.nodes ?? [];
		const accessManagerEmails = new Set(
			$store.data?.team.accessManagers.map((am) => am.user.email) ?? []
		);
		return allUsers.filter(
			(user) =>
				!accessManagerEmails.has(user.email) &&
				user.email !== $store.data?.team.section.manager?.email
		);
	});

	let email: string = $state('');

	let errors: string[] = $state([]);
	const submit = async () => {
		errors = [];
		const userID = $store.data?.users.nodes.find((u) => u.email === email)?.email;
		if (!userID) {
			errors = ['Fant ikke brukeren'];
			return;
		}

		const input: AddTeamAccessManagerInput = {
			teamSlug: team,
			userEmail: userID
		};

		const resp = await create.mutate({
			input
		});

		if (resp.errors) {
			errors = resp.errors
				.filter((e: { message: string }) => e.message != 'unable to resolve')
				.map((e: { message: string }) => e.message);
			return;
		}

		open = false;
		email = '';

		oncreated?.();
	};
</script>

<Modal bind:open>
	{#snippet header()}
		<Heading>Legg til tilgangsansvarlig</Heading>
	{/snippet}

	{#each errors as error (error)}
		<Alert variant="error">{error}</Alert>
	{/each}

	<form
		onsubmit={(e: SubmitEvent) => {
			e.preventDefault();
			submit();
		}}
		class="wrapper"
	>
		<TextField list="add-member-email" type="email" bind:value={email}>
			{#snippet label()}
				E-post
			{/snippet}
		</TextField>
		<datalist id="add-member-email">
			{#each users as user (user.email)}
				<option value={user.email}>{user.name}</option>
			{/each}
		</datalist>
	</form>

	{#snippet footer()}
		<Button type="submit" onclick={submit} icon={PlusIcon}>Legg til</Button>
	{/snippet}
</Modal>

<style>
	.wrapper {
		min-width: 400px;
	}

	:global(.tooltipAddMemberWrapper) {
		width: 200px;
	}
</style>
