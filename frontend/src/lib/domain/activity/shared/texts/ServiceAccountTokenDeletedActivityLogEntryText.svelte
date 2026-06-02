<script lang="ts">
	import Time from '$lib/ui/Time.svelte';
	import { BodyShort } from '@nais/ds-svelte-community';
	import type { ActivityLogEntry } from './types';

	let {
		data
	}: {
		data: ActivityLogEntry<'ServiceAccountTokenDeletedActivityLogEntry'>;
	} = $props();
</script>

<div>
	Nøkkel <i>{data.data.tokenName}</i> for tjenestekonto <i>{data.resourceName}</i> ble slettet
	<BodyShort size="small" textColor="subtle">
		av {data.actor} for
		<Time time={data.createdAt} distance />
	</BodyShort>
</div>
