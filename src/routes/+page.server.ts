export async function load(event) {
	const teamTableFields = JSON.parse(event.cookies.get('daplaTableFields') ?? '[]');

	return {
		teamTableFields: teamTableFields
	};
}
