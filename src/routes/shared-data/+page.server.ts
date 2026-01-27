export async function load(event) {
	const bucketTableFields = JSON.parse(event.cookies.get('daplaTableFieldsshared-data') ?? '[]');

	return {
		bucketTableFields: bucketTableFields
	};
}
