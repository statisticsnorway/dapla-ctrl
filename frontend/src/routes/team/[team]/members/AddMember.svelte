<script lang="ts">
	import { graphql, type AddGroupMemberInput } from '$houdini';
	import {
		Alert,
		Button,
		Checkbox,
		CheckboxGroup,
		Heading,
		Modal,
		TextField
	} from '@nais/ds-svelte-community';
	import { PlusIcon } from '@nais/ds-svelte-community/icons';

	interface Props {
		open: boolean;
		team: string;
		groups: { id: string; name: string }[];
		oncreated?: () => void;
	}

	let { open = $bindable(), team, groups, oncreated }: Props = $props();

	let selectedGroups: string[] = $state([]);

	const store = graphql(`
		query AddMemberQuery($team: Slug!) {
			users(first: 10000) {
				nodes {
					id
					email
					name
				}
			}
			team(slug: $team) {
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

	$effect(() => {
		store.fetch({
			variables: {
				team: team
			}
		});
	});

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

	let users = $derived.by(() => {
		const allUsers = $store.data?.users.nodes ?? [];
		const groupMemberEmails = new Set(
			$store.data?.team.members.nodes.map((member) => member.user.email) ?? []
		);
		return allUsers.filter((user) => !groupMemberEmails.has(user.email));
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

		for (const group of selectedGroups) {
			const input: AddGroupMemberInput = {
				groupName: group,
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
		}

		open = false;
		email = '';

		oncreated?.();
	};
</script>

<Modal bind:open>
	{#snippet header()}
		<Heading>Legg til medlem</Heading>
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
		<br />
		<datalist id="add-member-email">
			{#each users as user (user.email)}
				<option value={user.email}>{user.name}</option>
			{/each}
		</datalist>
		<CheckboxGroup legend="Grupper" bind:value={selectedGroups}>
			{#each groups as group (group.name)}
				<Checkbox value={group.name}>{group.name.substring(team.length + 1)}</Checkbox>
			{/each}
		</CheckboxGroup>
	</form>

	{#snippet footer()}
		<Button type="submit" onclick={submit} icon={PlusIcon}>Legg til medlem</Button>
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
