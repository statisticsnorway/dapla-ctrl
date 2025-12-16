<script lang="ts">
	import { graphql, type CreateGroupInput } from '$houdini';
	import { Alert, Button, Heading, Modal, Select, TextField } from '@nais/ds-svelte-community';
	import { PlusIcon } from '@nais/ds-svelte-community/icons';
	import { createEventDispatcher } from 'svelte';

	interface Props {
		open: boolean;
		team: string;
	}

	let { open = $bindable(), team }: Props = $props();

	let groupSuffix: string = $state('');
	let groupCategory: string = $state('developers');

	const dispatcher = createEventDispatcher<{ created: null }>();

	const create = graphql(`
		mutation CreateGroupMutation($input: CreateGroupInput!) {
			createGroup(input: $input) {
				group {
					name
				}
			}
		}
	`);

	let errors: string[] = $state([]);
	const submit = async () => {
		errors = [];

		const input: CreateGroupInput = {
			teamSlug: team,
			category: groupCategory,
			suffix: groupSuffix
		};

		const resp = await create.mutate({
			input
		});

		if (resp.errors) {
			errors = resp.errors.filter((e) => e.message != 'unable to resolve').map((e) => e.message);
			return;
		}

		open = false;

		dispatcher('created', null);
	};
</script>

<Modal bind:open>
	{#snippet header()}
		<Heading>Opprett gruppe for {team}</Heading>
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
		<Select label="Kategori" bind:value={groupCategory}>
			<option value="data-admins">data-admins</option>
			<option value="developers">developers</option>
		</Select>
		<TextField list="create-group-suffix" type="text" bind:value={groupSuffix}>
			{#snippet label()}
				Suffiks (for egendefinerte grupper)
			{/snippet}
		</TextField>
	</form>

	{#snippet footer()}
		<Button type="submit" onclick={submit} icon={PlusIcon}>Opprett gruppe</Button>
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
