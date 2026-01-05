export async function load(event) {
	const teamTableFields = JSON.parse(event.cookies.get('teamTableFieldsteams') ?? '[]');

	return {
		teamTableFields: teamTableFields
	};
}
