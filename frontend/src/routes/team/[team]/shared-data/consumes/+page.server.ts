export async function load(event) {
	const bucketTableFields = JSON.parse(
		event.cookies.get('consumesSharedBucketsFields/team') ?? '[]'
	);

	return {
		bucketTableFields: bucketTableFields
	};
}
