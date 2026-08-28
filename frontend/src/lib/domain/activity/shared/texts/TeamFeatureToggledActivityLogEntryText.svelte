<script lang="ts">
	import Time from '$lib/ui/Time.svelte';
	import { capitalizeFirstLetter } from '$lib/utils/formatters';
	import type { ActivityLogEntry } from './types';
	import { BodyShort } from '@nais/ds-svelte-community';

	let {
		data
	}: {
		data: ActivityLogEntry<
			'TeamFeatureEnabledActivityLogEntry' | 'TeamFeatureDisabledActivityLogEntry'
		>;
	} = $props();

	const featureToDisplayName = (feature: string) => {
		switch (feature) {
			case 'ai':
				return 'KI-funksjonalitet';
			default:
				return capitalizeFirstLetter(feature);
		}
	};
</script>

<div>
	{featureToDisplayName(data.resourceName)} ble slått
	{data.__typename === 'TeamFeatureEnabledActivityLogEntry' ? 'på' : 'av'}
	i {data.env}

	<BodyShort textColor="subtle" size="small">
		av {data.actor} for
		<Time time={data.createdAt} distance />
	</BodyShort>
</div>
