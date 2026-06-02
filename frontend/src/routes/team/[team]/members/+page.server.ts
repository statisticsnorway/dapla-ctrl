export async function load(event) {
	const groupMemberTableFields = JSON.parse(event.cookies.get('teamMembersTableFields') ?? '[]');

	return {
		groupMemberTableFields: groupMemberTableFields
	};
}
