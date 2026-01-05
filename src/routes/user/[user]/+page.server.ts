export async function load(event) {
	const teamTableFields = JSON.parse(event.cookies.get('teamTableFieldsuser') ?? '[]');

	return {
		teamTableFields: teamTableFields
	};
}
