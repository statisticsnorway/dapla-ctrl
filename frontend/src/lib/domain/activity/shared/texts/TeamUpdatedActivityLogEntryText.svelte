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
	const fieldValueToDisplayValue = (fieldName: string, value: string | null | undefined) => {
		if (fieldName === 'hasManualEditing') {
			if (value === 'true') return 'aktivert';
			if (value === 'false') return 'deaktivert';
		}
		return value;
	};
</script>

<div>
	{#if data.teamUpdated?.updatedFields.length}
		{#each data.teamUpdated?.updatedFields as field (field)}
			{fieldNameToDisplayName(field.field)} endret fra
			<i>{fieldValueToDisplayValue(field.field, field.oldValue)}</i> til
			<i>{fieldValueToDisplayValue(field.field, field.newValue)}</i>
		{/each}
	{/if}

	<BodyShort textColor="subtle" size="small">
		av {data.actor} for
		<Time time={data.createdAt} distance />
	</BodyShort>
</div>
