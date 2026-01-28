<script lang="ts">
	import Time from '$lib/ui/Time.svelte';
	import type { ActivityLogEntry } from './types';
	import { BodyShort } from '@nais/ds-svelte-community';

	let {
		data
	}: {
		data: ActivityLogEntry<'TeamUpdatedActivityLogEntry'>;
	} = $props();

	const fieldNameToDisplayName = (fieldName: string) => {
		switch (fieldName) {
			case 'displayName':
				return 'Visningsnavn';
			default:
				return fieldName;
		}
	};
</script>

<div>
	{#if data.teamUpdated?.updatedFields.length}
		{#each data.teamUpdated?.updatedFields as field (field)}
			{fieldNameToDisplayName(field.field)} endret fra <i>{field.oldValue}</i> til
			<i>{field.newValue}</i>
		{/each}
	{/if}

	<BodyShort textColor="subtle" size="small">
		av {data.actor} for
		<Time time={data.createdAt} distance />
	</BodyShort>
</div>
