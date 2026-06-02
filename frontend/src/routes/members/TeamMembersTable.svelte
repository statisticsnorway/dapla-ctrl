<script lang="ts">
	import Pagination from '$lib/ui/Pagination.svelte';
	import DaplaTable from '$lib/ui/DaplaTable.svelte';
	import { UserOrderField } from '$houdini';

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

	export interface TeamMemberData {
		id: string;
		user: {
			name: string;
			email: string;
		};
		teamCount: number;
		dataAdminCount: number;
		section?: {
			name: string;
			code: string;
		};
	}

	let { teamMembers, pageInfo, loaders, selected }: Props = $props();
</script>

{#snippet nameCell(member: TeamMemberData)}
	<a href="/member/{member.user.email}">
		<b>{member.user.name}</b>
	</a>
	<br />
	{member.user.email}
	<br />
	{#if member.section}
		{member.section.name} ({member.section.code})
	{:else}
		<span style="color: var(--ax-text-subtle); font-style: italic;">Mangler seksjon</span>
	{/if}
{/snippet}
{#snippet teamCell(member: TeamMemberData)}
	{member.teamCount}
{/snippet}
{#snippet dataAdminCell(member: TeamMemberData)}
	{member.dataAdminCount}
{/snippet}

<DaplaTable
	data={teamMembers}
	{selected}
	columns={[
		{
			id: 'NAME',
			name: 'Navn',
			show: 'ALWAYS',
			cell: nameCell,
			sortKey: UserOrderField.NAME
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
		}
	]}
/>

{#if pageInfo && loaders}
	<Pagination page={pageInfo} {loaders} />
{/if}
