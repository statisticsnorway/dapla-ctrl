export async function load(event) {
	const bucketTableFields = JSON.parse(event.cookies.get('sharedBucketsTableFields/user') ?? '[]');

	return {
		bucketTableFields: bucketTableFields
	};
}
