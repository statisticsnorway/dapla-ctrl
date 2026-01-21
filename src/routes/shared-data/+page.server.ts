export async function load(event) {
	const bucketTableFields = JSON.parse(event.cookies.get('bucketTableFieldsshared-data') ?? '[]');

	return {
		bucketTableFields: bucketTableFields
	};
}
