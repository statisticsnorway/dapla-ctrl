<script lang="ts">
	import { page } from '$app/state';
	import { menuItems } from '$lib/menuItems';
	import Menu from '$lib/ui/Menu.svelte';

	const {
		member,
		isAdmin,
		features,
		teamSlug
	}: {
		member: boolean;
		isAdmin: boolean;
		features?: object;
		teamSlug: string;
	} = $props();
</script>

<Menu
	items={menuItems({
		path: page.url.pathname,
		member,
		features,
		isAdmin
	})}
/>
