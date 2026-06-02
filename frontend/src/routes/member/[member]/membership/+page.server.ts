export async function load(event) {
	const teamTableFields = JSON.parse(
		event.cookies.get('daplaTableFields/member/membership') ?? '[]'
	);

	return {
		teamTableFields: teamTableFields
	};
}
