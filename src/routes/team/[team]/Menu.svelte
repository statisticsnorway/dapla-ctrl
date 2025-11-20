<script lang="ts">
	import { page } from '$app/state';
	import { graphql } from '$houdini';
	import Menu from '$lib/components/Menu.svelte';
	import { menuItems } from '$lib/menuItems';

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
