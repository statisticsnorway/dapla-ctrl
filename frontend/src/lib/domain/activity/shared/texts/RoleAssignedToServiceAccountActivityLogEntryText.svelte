<script lang="ts">
	import Time from '$lib/ui/Time.svelte';
	import { BodyShort } from '@nais/ds-svelte-community';
	import type { ActivityLogEntry } from './types';

	let {
		data
	}: {
		data: ActivityLogEntry<'RoleAssignedToServiceAccountActivityLogEntry'>;
	} = $props();
</script>

<div>
	Tjenestekonto <i>{data.resourceName}</i> ble gitt rollen <i>{data.data.roleName}</i>
	<BodyShort size="small" textColor="subtle">
		av {data.actor} for
		<Time time={data.createdAt} distance />
	</BodyShort>
</div>
