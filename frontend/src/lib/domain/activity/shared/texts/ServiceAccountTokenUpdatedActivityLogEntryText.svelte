<script lang="ts">
	import Time from '$lib/ui/Time.svelte';
	import { BodyShort } from '@nais/ds-svelte-community';
	import type { ActivityLogEntry } from './types';

	let {
		data
	}: {
		data: ActivityLogEntry<'ServiceAccountTokenUpdatedActivityLogEntry'>;
	} = $props();
</script>

<div>
	Nøkkel for tjenestekonto <i>{data.resourceName}</i> ble oppdatert
	{#if data.data.updatedFields.length}
		<br />
		<BodyShort size="small" textColor="subtle">
			{#each data.data.updatedFields as field (field.field)}
				{field.field} endret fra <i>{field.oldValue}</i> til <i>{field.newValue}</i>
			{/each}
		</BodyShort>
	{/if}
	<BodyShort size="small" textColor="subtle">
		av {data.actor} for
		<Time time={data.createdAt} distance />
	</BodyShort>
</div>
