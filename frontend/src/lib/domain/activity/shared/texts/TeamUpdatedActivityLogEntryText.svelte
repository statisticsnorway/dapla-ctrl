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
			case 'sectionCode':
				return 'Seksjonskode';
			case 'hasManualEditing':
				return 'Parquedit';
			default:
				return fieldName;
		}
	};
	const updatedFieldToDisplayText = (fieldName: string, value: string | null | undefined) => {
		if (fieldName === 'hasManualEditing') {
		    return value === 'true'
				? 'Parquedit ble skrudd på'
				: 'Parquedit ble skrudd av';
		}
		return value;
	};
</script>

<div>
	{#each data.teamUpdated?.updatedFields as field(field)}
		{#if updatedFieldToDisplayText(field.field, field.newValue)}
		    {updatedFieldToDisplayText(field.field, field.newValue)}
		{:else}
			{fieldNameToDisplayName(field.field)} endret fra <i>(field.oldValue</i> til<i>field.newValue</i>
		{/if}
	{/each}

	<BodyShort textColor="subtle" size="small">
		av {data.actor} for
		<Time time={data.createdAt} distance />
	</BodyShort>
</div>
