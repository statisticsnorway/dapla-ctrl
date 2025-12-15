<script lang="ts">
	import Pagination from '$lib/ui/Pagination.svelte';
	import { BodyShort, Heading } from '@nais/ds-svelte-community';
	import type { PageProps } from './$types';
	import TeamsTable, { type TeamsData } from '../../TeamsTable.svelte';
	import type { UserOverview$result } from '$houdini';

	let { data }: PageProps = $props();

	let { UserOverview } = $derived(data);

	type TeamNode = UserOverview$result['user']['teams']['nodes'][0];

	let allTeamsCount = $derived($UserOverview.data?.user.teams.nodes.length || 0);
	function transformTeamData(teamNode: TeamNode): TeamsData {
		const team = teamNode.team;
		const manager = team.section.manager;

		return {
			slug: team.slug,
			purpose: team.purpose,
			memberCount: team.members.pageInfo.totalCount,
			manager: {
				name: manager?.name ?? 'Mangler seksjonsleder',
				email: manager?.email ?? ''
			},
			section: {
				code: team.section.code,
				name: team.section.name
			}
		};
	}
</script>

{#if $UserOverview.data}
	<div class="user-info">
		<div>
			<BodyShort>{$UserOverview.data.user.email}</BodyShort>
		</div>
	</div>

	<div class="container">
		<div>
			<div class="section-header">
				<Heading level="2" spacing>Medlem av {allTeamsCount} team</Heading>
			</div>
			<TeamsTable teamsData={$UserOverview.data.user.teams.nodes.map(transformTeamData)} />
		</div>
	</div>
	<Pagination
		page={$UserOverview.data.user.teams.pageInfo}
		loaders={{
			loadPreviousPage: () => UserOverview.loadPreviousPage(),
			loadNextPage: () => UserOverview.loadNextPage()
		}}
	/>
{/if}

<style>
	.user-info {
		display: grid;
		grid-template-columns: 1fr 300px;
		gap: var(--spacing-layout);
		margin-top: calc(-1 * var(--spacing-layout));
		margin-bottom: var(--spacing-layout);
	}
</style>
