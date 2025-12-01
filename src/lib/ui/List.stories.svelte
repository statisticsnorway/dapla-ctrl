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
			slug: 'aura',
			purpose: 'Not the Nais platform team any more, check out Nais instead.'
		},
		{
			slug: 'detsombetyrnoe',
			purpose:
				'Rekruttering og profileringssiden for NAV IT - blir liggende på detsombetyrnoe.no. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Ut elit tellus, luctus nec ullamcorper mattis, pulvinar dapibus leo.'
		},
		{
			slug: 'frontend-golden-path',
			purpose: 'Teste greier for den gyldne sti'
		},
		{
			slug: 'frontendplattform',
			purpose: 'Frontendplattform'
		},
		{
			slug: 'nais',
			purpose: 'The Nais platform team'
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
