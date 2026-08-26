<script lang="ts">
	import Time from '$lib/ui/Time.svelte';
	import type { ActivityLogEntry } from './types';
	import { BodyShort } from '@nais/ds-svelte-community';

	let {
		data
	}: {
		data: ActivityLogEntry<'TeamFeatureDisabledActivityLogEntry'>;
	} = $props();

	const featureToDisplayName = (feature: string) => {
		switch (feature) {
			case 'ai':
				return 'AI-funksjonalitet';
			default:
				return 'Ukjent funksjonalitet';
		}
	};
</script>

<div>
	{featureToDisplayName(data.resourceName)} ble slått av i {data.env}

	<BodyShort textColor="subtle" size="small">
		av {data.actor} for
		<Time time={data.createdAt} distance />
	</BodyShort>
</div>
