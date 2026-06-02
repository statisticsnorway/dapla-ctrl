<script lang="ts">
	import Time from '$lib/ui/Time.svelte';
	import { BodyShort } from '@nais/ds-svelte-community';
	import type { ActivityLogEntry } from './types';

	let {
		data
	}: {
		data: ActivityLogEntry<'ReconcilerDisabledActivityLogEntry'>;
	} = $props();
</script>

<div>
	Reconciler <i>{data.resourceName}</i> ble deaktivert
	<BodyShort size="small" textColor="subtle">
		av {data.actor} for
		<Time time={data.createdAt} distance />
	</BodyShort>
</div>
