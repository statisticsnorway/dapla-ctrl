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
			case 'purpose':
				return 'beskrivelse';
			case 'displayName':
				return 'visningsnavn';
			default:
				return fieldName;
		}
	};
</script>

<div>
	Oppdaterte team
	{#if data.teamUpdated?.updatedFields.length}
		{#each data.teamUpdated?.updatedFields as field (field)}
			{fieldNameToDisplayName(field.field)}. Endret fra <i>{field.oldValue}</i> til
			<i>{field.newValue}</i>.
		{/each}
	{/if}

	<BodyShort textColor="subtle" size="small">
		av {data.actor}
		<Time time={data.createdAt} distance />
	</BodyShort>
</div>
