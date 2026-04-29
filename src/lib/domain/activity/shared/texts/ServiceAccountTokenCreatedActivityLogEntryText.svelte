<script lang="ts">
	import Time from '$lib/ui/Time.svelte';
	import { BodyShort } from '@nais/ds-svelte-community';
	import type { ActivityLogEntry } from './types';

	let {
		data
	}: {
		data: ActivityLogEntry<'ServiceAccountTokenCreatedActivityLogEntry'>;
	} = $props();
</script>

<div>
	Nøkkel for tjenestekonto <i>{data.resourceName}</i> ble opprettet
	<BodyShort size="small" textColor="subtle">
		av {data.actor} for
		<Time time={data.createdAt} distance />
	</BodyShort>
</div>
