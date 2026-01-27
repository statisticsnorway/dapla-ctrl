export async function load(event) {
	const bucketTableFields = JSON.parse(event.cookies.get('sharedBucketUsersFields/team') ?? '[]');

	return {
		bucketTableFields: bucketTableFields
	};
}
