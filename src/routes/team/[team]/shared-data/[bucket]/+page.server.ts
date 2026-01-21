export async function load(event) {
	const bucketTableFields = JSON.parse(event.cookies.get('bucketUsersTableFieldsteam') ?? '[]');

	return {
		bucketTableFields: bucketTableFields
	};
}
