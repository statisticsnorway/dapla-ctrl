<script lang="ts">
	import Time from '$lib/ui/Time.svelte';
	import { BodyShort } from '@nais/ds-svelte-community';
	import type { ActivityLogEntry } from './types';

	let {
		data
	}: {
		data: ActivityLogEntry<'ReconcilerConfiguredActivityLogEntry'>;
	} = $props();
</script>

<div>
	Reconciler <i>{data.resourceName}</i> ble konfigurert
	{#if data.data.updatedKeys.length}
		<br />
		<BodyShort size="small" textColor="subtle">
			Oppdaterte nøkler: {data.data.updatedKeys.join(', ')}
		</BodyShort>
	{/if}
	<BodyShort size="small" textColor="subtle">
		av {data.actor} for
		<Time time={data.createdAt} distance />
	</BodyShort>
</div>
