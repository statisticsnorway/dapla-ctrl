<script lang="ts">
	import Time from '$lib/ui/Time.svelte';
	import { BodyShort } from '@nais/ds-svelte-community';
	import type { ActivityLogEntry } from './types';

	let {
		data
	}: {
		data: ActivityLogEntry<'TeamRoleRevokedActivityLogEntry'>;
	} = $props();
</script>

<div class="container">
	{data.roleRevoked.user.email} ble fratatt rollen <i>{data.roleRevoked.role}</i>
	<BodyShort textColor="subtle" size="small">
		av {data.actor} for
		<Time time={data.createdAt} distance />
	</BodyShort>
</div>

<style>
	.container {
		overflow-wrap: normal;
		word-break: normal;
		line-height: 1.4;
	}
</style>
