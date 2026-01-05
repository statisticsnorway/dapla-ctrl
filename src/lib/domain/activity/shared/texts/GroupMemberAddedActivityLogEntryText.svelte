<script lang="ts">
	import Time from '$lib/ui/Time.svelte';
	import { BodyShort } from '@nais/ds-svelte-community';
	import type { ActivityLogEntry } from './types';

	let {
		data
	}: {
		data: ActivityLogEntry<'GroupMemberAddedActivityLogEntry'>;
	} = $props();
</script>

<div>
	{data.groupMemberAdded.userEmail} ble lagt til i gruppe
	<BodyShort textColor="subtle" size="small">
		av {data.actor}
		<Time time={data.createdAt} distance />
	</BodyShort>
</div>
