/// <references types="houdini-svelte">

/** @type {import('houdini').ConfigFile} */

const graphqlEndpoint = import.meta.env.VITE_GRAPHQL_ENDPOINT;
if (!graphqlEndpoint) {
	console.log(`env variable 'VITE_GRAPHQL_ENDPOINT' must be set`);
	process.exit(1);
}

const config = {
	schemaPath: './schema.graphql',
	url: graphqlEndpoint,
	runtimeDir: '.houdini',
	defaultPaginateMode: 'SinglePage',
	watchSchema: {
		interval: 0,
		url: 'env:VITE_SCHEMA_ENDPOINT',
		headers: {
			'x-user-email': 'dev.usersen@example.com'
		}
	},
	defaultCachePolicy: 'CacheAndNetwork',
	plugins: {
		'houdini-svelte': {}
	},
	scalars: {
		Slug: { type: 'string' },
		Cursor: { type: 'string' },
		Date: {
			type: 'Date',
			unmarshal(val) {
				return new Date(val);
			},
			// turn the value into something the API can use
			marshal(d) {
				if (typeof d === 'string') {
					return d;
				}
				return (
					d.getFullYear() +
					'-' +
					('0' + (d.getMonth() + 1)).slice(-2) +
					'-' +
					('0' + d.getDate()).slice(-2)
				);
			}
		},
		Time: {
			type: 'Date',
			unmarshal(val) {
				return new Date(val);
			},
			// turn the value into something the API can use
			marshal(date) {
				return date.toISOString();
			}
		},
		TimeOfDay: { type: 'string' }
	}
};

export default config;
