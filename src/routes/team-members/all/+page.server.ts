export async function load(event) {
	const teamMembersTableFields = JSON.parse(
		event.cookies.get('daplaTableFieldsteam-members') ?? '[]'
	);

	return {
		teamMembersTableField: teamMembersTableFields
	};
}
