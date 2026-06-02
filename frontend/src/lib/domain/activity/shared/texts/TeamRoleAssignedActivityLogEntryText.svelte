<script lang="ts">
	import Time from '$lib/ui/Time.svelte';
	import { BodyShort } from '@nais/ds-svelte-community';
	import type { ActivityLogEntry } from './types';

	let {
		data
	}: {
		data: ActivityLogEntry<'TeamRoleAssignedActivityLogEntry'>;
	} = $props();
</script>

<div>
	{data.roleAssigned.user.email} ble gitt rollen <i>{data.roleAssigned.role}</i>
	<BodyShort textColor="subtle" size="small">
		av {data.actor} for
		<Time time={data.createdAt} distance />
	</BodyShort>
</div>
