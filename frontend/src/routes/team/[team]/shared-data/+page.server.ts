export async function load(event) {
	const bucketTableFields = JSON.parse(event.cookies.get('sharedBucketsFields/team') ?? '[]');

	return {
		bucketTableFields: bucketTableFields
	};
}
