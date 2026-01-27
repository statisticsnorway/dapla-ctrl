export async function load(event) {
	const teamTableFields = JSON.parse(event.cookies.get('daplaTableFieldsteams') ?? '[]');

	return {
		teamTableFields: teamTableFields
	};
}
