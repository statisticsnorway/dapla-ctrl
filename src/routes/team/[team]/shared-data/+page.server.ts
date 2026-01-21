export async function load(event) {
	const bucketTableFields = JSON.parse(event.cookies.get('bucketTableFieldsteam') ?? '[]');

	return {
		bucketTableFields: bucketTableFields
	};
}
