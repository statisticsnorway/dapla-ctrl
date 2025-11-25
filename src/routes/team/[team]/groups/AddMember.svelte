<script lang="ts">
	import { graphql, type AddGroupMemberInput } from '$houdini';
	import { Alert, Button, Heading, Modal, Select, TextField } from '@nais/ds-svelte-community';
	import { PlusIcon } from '@nais/ds-svelte-community/icons';
	import { createEventDispatcher } from 'svelte';
	import type { AddMemberQueryVariables } from './$houdini';

	interface Props {
		open: boolean;
		groups: string[];
	}

	let { open = $bindable(), groups }: Props = $props();

	let group = $derived(groups[0]);

	const dispatcher = createEventDispatcher<{ created: null }>();

	export const _AddMemberQueryVariables: AddMemberQueryVariables = () => {
		return { group: group };
	};

	const store = graphql(`
		query AddMemberQuery($group: String!) @load {
			users(first: 10000) {
				nodes {
					id
					email
				}
			}
			group(name: $group) {
				members {
					nodes {
						user {
							email
						}
					}
				}
			}
		}
	`);

	const create = graphql(`
		mutation CreateMemberMutation($input: AddGroupMemberInput!) {
			addGroupMember(input: $input) {
				member {
					group {
						members {
							nodes {
								user {
									email
								}
							}
						}
					}
					user {
						id
					}
				}
			}
		}
	`);

	let emails = $derived.by(() => {
		const allEmails = $store.data?.users.nodes.map((user) => user.email) ?? [];
		const groupMemberEmails = new Set(
			$store.data?.group.members.nodes.map((member) => member.user.email) ?? []
		);
		return allEmails.filter((email) => !groupMemberEmails.has(email));
	});

	let email: string = $state('');

	let errors: string[] = $state([]);
	const submit = async () => {
		errors = [];
		const userID = $store.data?.users.nodes.find((u) => u.email === email)?.email;
		if (!userID) {
			errors = ['User not found'];
			return;
		}

		const input: AddGroupMemberInput = {
			groupName: group,
			userEmail: userID
		};

		const resp = await create.mutate({
			input
		});

		if (resp.errors) {
			errors = resp.errors.filter((e) => e.message != 'unable to resolve').map((e) => e.message);
			return;
		}

		open = false;
		email = '';

		dispatcher('created', null);
	};
</script>

<Modal bind:open>
	{#snippet header()}
		<Heading>Add Member</Heading>
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
		<p>Group members are given access to the some of the team's resources.</p>
		<Select label="Group" bind:value={group}>
			{#each groups as groupName (groupName)}
				<option value={groupName}>{groupName}</option>
			{/each}
		</Select>
		<TextField list="add-member-email" type="email" bind:value={email}>
			{#snippet label()}
				Email
			{/snippet}
		</TextField>
		<datalist id="add-member-email">
			{#each emails as email (email)}
				<option value={email}>{email}</option>
			{/each}
		</datalist>
	</form>

	{#snippet footer()}
		<Button type="submit" onclick={submit} icon={PlusIcon}>Add member</Button>
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
