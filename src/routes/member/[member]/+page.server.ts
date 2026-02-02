export async function load(event) {
	const teamTableFields = JSON.parse(event.cookies.get('daplaTableFields/member') ?? '[]');

	return {
		teamTableFields: teamTableFields
	};
}
