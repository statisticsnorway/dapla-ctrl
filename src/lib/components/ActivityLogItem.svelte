<script lang="ts">
	import {
		ActivityLogEntryResourceType,
		fragment,
		graphql,
		type ActivityLogEntryFragment,
		type ActivityLogEntryResourceType$options
	} from '$houdini';
	import Time from '$lib/Time.svelte';
	import { BodyShort } from '@nais/ds-svelte-community';

	const resourceLink = (
		resourceType: ActivityLogEntryResourceType$options,
		resourceName: string,
		teamSlug: string | null
	) => {
		switch (resourceType) {
			case ActivityLogEntryResourceType.TEAM:
				return `/team/${teamSlug}`;
			default:
				return null;
		}
	};

	interface Props {
		item: ActivityLogEntryFragment;
	}

	let { item }: Props = $props();

	let data = $derived(
		fragment(
			item,
			graphql(`
				fragment ActivityLogEntryFragment on ActivityLogEntry {
					__typename
					createdAt
					actor
					createdAt
					message
					resourceName
					resourceType
					teamSlug
				}
			`)
		)
	);
</script>

<div class="activity">
	<div>
		<BodyShort size="small" spacing>
			{#if $data.__typename === 'TeamMemberAddedActivityLogEntry'}
				{#if $data.teamMemberAdded}
					Added member {$data.teamMemberAdded.userEmail !== ''
						? $data.teamMemberAdded.userEmail
						: 'unknown email'} to team as {$data.teamMemberAdded.role}.
				{/if}
			{:else if $data.__typename === 'TeamMemberRemovedActivityLogEntry'}
				{#if $data.teamMemberRemoved}
					Removed <strong
						>{$data.teamMemberRemoved.userEmail !== ''
							? $data.teamMemberRemoved.userEmail
							: 'unknown email'}</strong
					>
					from team.
				{/if}
			{:else if $data.__typename === 'TeamMemberSetRoleActivityLogEntry'}
				{#if $data.teamMemberSetRole}
					Set role to <strong>{$data.teamMemberSetRole.role}</strong> for user {$data
						.teamMemberSetRole.userEmail !== ''
						? $data.teamMemberSetRole.userEmail
						: 'unknown email'}.
				{/if}
			{:else if $data.__typename === 'TeamUpdatedActivityLogEntry'}
				{$data.message}
				{#if $data.teamUpdated?.updatedFields.length}
					{#each $data.teamUpdated?.updatedFields as field (field)}
						{field.field}. Changed from <i>{field.oldValue}</i> to <i>{field.newValue}</i>.
					{/each}
				{/if}
			{:else if $data.__typename === 'UnleashInstanceUpdatedActivityLogEntry'}
				{@const u = $data.unleashInstanceUpdated}
				{$data.message}
				{#if u.allowedTeamSlug}
					Allowed <a href="/team/{u.allowedTeamSlug}">
						{u.allowedTeamSlug}
					</a> to access the instance.
				{:else if u.revokedTeamSlug}
					Revoked access for <a href="/team/{u.revokedTeamSlug}">
						{u.revokedTeamSlug}
					</a> to the instance.
				{/if}
			{:else if $data.__typename === 'RepositoryRemovedActivityLogEntry'}
				<a href="/team/{$data.teamSlug}/repositories">Repository</a>
				<strong>{$data.resourceName}</strong> removed from team {$data.teamSlug}.
			{:else if $data.__typename === 'RepositoryAddedActivityLogEntry'}
				<a href="/team/{$data.teamSlug}/repositories">Repository</a>
				<strong>{$data.resourceName}</strong> added to team {$data.teamSlug}.
			{:else if $data.__typename === 'ApplicationDeletedActivityLogEntry'}
				Application <strong>{$data.resourceName}</strong> was deleted
			{:else if $data.__typename === 'ApplicationRestartedActivityLogEntry'}
				Application <strong>{$data.resourceName}</strong> was restarted
			{:else if $data.__typename === 'JobTriggeredActivityLogEntry'}
				Job <a href={resourceLink($data.resourceType, $data.resourceName, $data.teamSlug)}
					>{$data.resourceName}</a
				>
				was triggered
			{:else}
				{$data.message}
				{@const link = resourceLink($data.resourceType, $data.resourceName, $data.teamSlug)}
				{#if link}
					<a href={link}>{$data.resourceName}</a>
				{/if}
			{/if}
		</BodyShort>
	</div>
	<div>
		<BodyShort size="small" style="color: var(--ax-text-subtle, --a-text-subtle)">
			<Time time={$data.createdAt} distance={true} />
			by {$data.actor}
		</BodyShort>
	</div>
</div>

<style>
	.activity {
		display: flex;
		flex-direction: column;
	}
</style>
