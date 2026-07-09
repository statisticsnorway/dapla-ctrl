<script lang="ts">
	import Time from '$lib/ui/Time.svelte';
	import { BodyShort } from '@nais/ds-svelte-community';
	import type { ActivityLogEntryFragment$data } from '$houdini';
	let {
		data
	}: {
		data: ActivityLogEntryFragment$data;
	} = $props();
</script>

<div>
	{data.message}<br />
	<BodyShort textColor="subtle" size="small">
		av {data.actor} for
		<Time time={data.createdAt} distance />
	</BodyShort>
</div>
