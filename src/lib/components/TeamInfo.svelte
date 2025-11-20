<script lang="ts">
	import { page } from '$app/stores';
	import { graphql, PendingValue } from '$houdini';
	import { BodyShort, Heading, Skeleton } from '@nais/ds-svelte-community';
	import { get } from 'svelte/store';
	import type { TeamInfoVariables } from './$houdini';

	interface Props {
		teamSlug: string;
		viewerIsMember: boolean;
	}

	let { teamSlug, viewerIsMember }: Props = $props();

	export const _TeamInfoVariables: TeamInfoVariables = () => {
		return { team: teamSlug };
	};

	const teamInfo = graphql(`
		query TeamInfo($team: Slug!) @load {
			team(slug: $team) @loading {
				purpose @loading
			}
		}
	`);
	const githubOrganization = get(page).data.githubOrganization;
</script>

<div class="wrapper">
	<Heading level="4" size="small">Team Summary</Heading>

	{#if $teamInfo.data}
		{@const t = $teamInfo.data.team}
		{#if t.purpose !== PendingValue}
			<BodyShort>{t.purpose}</BodyShort>
		{:else}
			<Skeleton variant="text" />
		{/if}

		<BodyShort>
			{#if viewerIsMember}
				<a href="/team/{teamSlug}/settings">View team settings</a>
			{/if}
		</BodyShort>
	{/if}
</div>

<style>
	.wrapper {
		display: flex;
		flex-direction: column;
		gap: var(--ax-space-4, --a-spacing-1);
		align-items: start;
	}
</style>
