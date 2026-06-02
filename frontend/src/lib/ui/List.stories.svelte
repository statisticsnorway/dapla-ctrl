<script module>
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import List from '$lib/ui/List.svelte';

	const { Story } = defineMeta({
		component: List,
		tags: ['autodocs']
	});
</script>

<script lang="ts">
	import TeamListItem from '$lib/domain/list-items/TeamListItem.svelte';

	const teams = [
		{
			slug: 'aura'
		},
		{
			slug: 'detsombetyrnoe'
		},
		{
			slug: 'frontend-golden-path'
		},
		{
			slug: 'frontendplattform'
		},
		{
			slug: 'nais'
		}
	];
</script>

<Story name="Teams" asChild>
	<List>
		{#each teams as team (team.slug)}
			<TeamListItem {team} />
		{/each}
	</List>
</Story>
