export async function load(event) {
	const bucketTableFields = JSON.parse(event.cookies.get('bucketTableFieldsuser') ?? '[]');

	return {
		bucketTableFields: bucketTableFields
	};
}
