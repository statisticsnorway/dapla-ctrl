<script lang="ts">
	import {
		graphql,
		ModifyCreateMemberMutationStore,
		ModifyDestroyMemberMutationStore
	} from '$houdini';
	import {
		Alert,
		BodyShort,
		Button,
		Checkbox,
		CheckboxGroup,
		Heading,
		Label,
		Modal
	} from '@nais/ds-svelte-community';
	import { CheckmarkIcon, TrashIcon } from '@nais/ds-svelte-community/icons';

	interface Props {
		open: boolean;
		team: string;
		user: { name: string; email: string };
		groups: string[];
		currentGroups: string[];
		oncreated?: () => void;
	}

	let { open = $bindable(), team, groups, currentGroups, oncreated, user }: Props = $props();

	// We do not want this to be derived. We need a straight copy of the list,
	// as we want to change it and compare the changes.
	// svelte-ignore state_referenced_locally
	let selectedGroups = $state([...currentGroups]);

	const create = graphql(`
		mutation ModifyCreateMemberMutation($input: AddGroupMemberInput!) {
			addGroupMember(input: $input) {
				__typename
			}
		}
	`);

	const destroy = graphql(`
		mutation ModifyDestroyMemberMutation($input: RemoveGroupMemberInput!) {
			removeGroupMember(input: $input) {
				__typename
			}
		}
	`);

	// Adds or removes a member based on which action is passed
	const modifyMember = async (
		group: string,
		action: ModifyCreateMemberMutationStore | ModifyDestroyMemberMutationStore
	): Promise<string[] | undefined> => {
		const resp = await action.mutate({
			input: {
				groupName: group,
				userEmail: user.email
			}
		});

		return resp.errors
			?.filter((e: { message: string }) => e.message != 'unable to resolve')
			.map((e: { message: string }) => e.message);
	};

	let errors: string[] | undefined = $state();
	const submit = async () => {
		errors = undefined;

		// Add user to groups they are not a member of, but have selected
		for (const group of selectedGroups.filter((g) => !currentGroups.includes(g))) {
			errors = await modifyMember(group, create);

			if (errors) {
				return;
			}
		}

		// Remove user from groups they are a member of but have deselected
		for (const group of currentGroups.filter((g) => !selectedGroups.includes(g))) {
			errors = await modifyMember(group, destroy);

			if (errors) {
				return;
			}
		}

		open = false;

		oncreated?.();
	};
</script>

<Modal bind:open>
	{#snippet header()}
		<Heading>Endre medlem</Heading>
	{/snippet}

	{#each errors as error (error)}
		<Alert variant="error">{error}</Alert>
	{/each}

	<Label size="medium">Medlem</Label>
	<BodyShort>
		{user.name} ({user.email})
	</BodyShort>
	<br />
	<form
		onsubmit={(e: SubmitEvent) => {
			e.preventDefault();
			submit();
		}}
		class="wrapper"
	>
		<CheckboxGroup legend="Grupper" bind:value={selectedGroups}>
			{#each groups as grp (grp)}
				<Checkbox value={grp}>{grp.substring(team.length + 1)}</Checkbox>
			{/each}
		</CheckboxGroup>
	</form>

	{#snippet footer()}
		<Button type="submit" onclick={submit} icon={CheckmarkIcon}>Lagre</Button>
		<Button
			type="submit"
			variant="danger"
			onclick={() => {
				selectedGroups = [];
				submit();
			}}
			icon={TrashIcon}>Fjern fra teamet</Button
		>
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
