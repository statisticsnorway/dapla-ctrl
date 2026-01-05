<script lang="ts">
	import Time from '$lib/ui/Time.svelte';
	import { BodyShort } from '@nais/ds-svelte-community';
	import type { ActivityLogEntry } from './types';

	let {
		data
	}: {
		data: ActivityLogEntry<'GroupMemberRemovedActivityLogEntry'>;
	} = $props();
</script>

<div>
	{data.groupMemberRemoved.userEmail} ble fjernet fra gruppen
	<BodyShort textColor="subtle" size="small">
		av {data.actor}
		<Time time={data.createdAt} distance />
	</BodyShort>
</div>
