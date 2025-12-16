<script lang="ts">
	import { page } from '$app/state';
	import { menuItems } from '$lib/menuItems';
	import Menu from '$lib/ui/Menu.svelte';

	const {
		member,
		isAdmin
	}: {
		member: boolean;
		isAdmin: boolean;
	} = $props();
</script>

<Menu
	items={menuItems({
		path: page.url.pathname,
		member,
		isAdmin
	})}
/>
