export async function load(event) {
	const bucketTableFields = JSON.parse(
		event.cookies.get('sharedBucketsTableFields/member') ?? '[]'
	);

	return {
		bucketTableFields: bucketTableFields
	};
}
