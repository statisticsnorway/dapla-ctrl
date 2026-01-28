<script lang="ts">
	import { Table, Tbody, Td, Th, Thead, Tr } from '@nais/ds-svelte-community';

	interface Props {
		groupMembersData: GroupMembersData[];
	}

	let { groupMembersData }: Props = $props();

	export type GroupMembersData = {
		name: string;
		email: string;
		section: {
			code: string;
			name: string;
		};
		groups: {
			category: string;
			suffix: string | null;
		}[];
	};
</script>

<Table zebraStripes>
	<Thead>
		<Tr>
			<Th>Navn</Th>
			<Th>Grupper</Th>
		</Tr>
	</Thead>
	<Tbody>
		{#each groupMembersData as groupMember (groupMember.email)}
			<Tr shadeOnHover={false}>
				<Td>
					<a href={`/user/${groupMember.email}`}>
						<b>{groupMember.name}</b>
					</a>
					<br />
					{groupMember.email}
					<br />
					{groupMember.section.name} ({groupMember.section.code})
				</Td>
				<Td>
					{groupMember.groups
						.map((g) => (g.suffix && g.suffix !== '' ? `${g.category}-${g.suffix}` : g.category))
						.toSorted()
						.join(', ')}
				</Td>
			</Tr>
		{/each}
	</Tbody>
</Table>
