<script lang="ts">
	import Pagination from '$lib/ui/Pagination.svelte';
	import { type TeamMemberData } from './teamMembersUtils';
	import DaplaTable from '$lib/ui/DaplaTable.svelte';

	interface Props {
		teamMembers: TeamMemberData[];
		pageInfo?: {
			readonly hasNextPage: boolean;
			readonly hasPreviousPage: boolean;
			readonly pageStart: number;
			readonly pageEnd: number;
			readonly totalCount: number;
		};
		loaders?: {
			loadPreviousPage: () => void;
			loadNextPage: () => void;
		};
		selected: string[];
	}

	let { teamMembers, pageInfo, loaders, selected }: Props = $props();
</script>

{#snippet nameCell(member: TeamMemberData)}
	<a href="/user/{member.user.email}">{member.user.name}</a>
{/snippet}
{#snippet teamCell(member: TeamMemberData)}
	{member.teamCount}
{/snippet}
{#snippet dataAdminCell(member: TeamMemberData)}
	{member.dataAdminCount}
{/snippet}
{#snippet managerCell(member: TeamMemberData)}
	{#if member.sectionManager}
		{#if member.sectionManager.email}
			<a href="/user/{member.sectionManager.email}">{member.sectionManager.name}</a>
		{:else}
			{member.sectionManager.name}
		{/if}
	{:else}
		<span style="color: var(--ax-text-subtle); font-style: italic;">Mangler seksjonsleder</span>
	{/if}
{/snippet}

<DaplaTable
	data={teamMembers}
	{selected}
	columns={[
		{
			id: 'NAME',
			name: 'Navn',
			show: 'ALWAYS',
			cell: nameCell
		},
		{
			id: 'TEAM_COUNT',
			name: 'Team',
			align: 'right',
			show: 'DEFAULT_YES',
			cell: teamCell
		},
		{
			id: 'DATA_ADMIN_COUNT',
			name: 'Data-admins-roller',
			align: 'right',
			show: 'DEFAULT_YES',
			cell: dataAdminCell
		},
		{
			id: 'MANAGER',
			name: 'Seksjonsleder',
			show: 'DEFAULT_YES',
			cell: managerCell
		}
	]}
/>

{#if pageInfo && loaders}
	<Pagination page={pageInfo} {loaders} />
{/if}
