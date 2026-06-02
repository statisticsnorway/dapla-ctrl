export async function load(event) {
	const teamMembersTableFields = JSON.parse(event.cookies.get('daplaTableFieldsmembers') ?? '[]');

	return {
		teamMembersTableField: teamMembersTableFields
	};
}
