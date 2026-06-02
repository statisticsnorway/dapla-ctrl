<script lang="ts">
	import FavoritesListItem from '$lib/domain/list-items/FavoritesListItem.svelte';
	import { favorites } from '$lib/stores/favorites.svelte';
	import SortableList from '$lib/ui/SortableList.svelte';
	import { BodyLong, Heading } from '@nais/ds-svelte-community';
	import { StarIcon } from '@nais/ds-svelte-community/icons';
</script>

<div class="favorites-list">
	<div class="header">
		<Heading level="1" size="large">Favorittsider</Heading>
	</div>

	<SortableList
		items={favorites.getFavorites()}
		onReorder={(newOrder) => favorites.setFavorites(newOrder)}
	>
		{#each favorites.getFavorites() as fav (fav)}
			<FavoritesListItem path={fav} />
		{:else}
			<BodyLong>
				Du har ingen favoritter ennå. Trykk på <StarIcon /> for å legge til en side som favoritt.
			</BodyLong>
		{/each}
	</SortableList>
</div>

<style>
	.favorites-list {
		margin-top: var(--ax-space-16);
	}
	.header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--ax-space-16);
	}
</style>
