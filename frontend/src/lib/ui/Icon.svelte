<script lang="ts">
	import {
		FloppydiskIcon,
		HexagonGridIcon,
		PersonGroupIcon,
		QuestionmarkIcon,
		ShieldLockIcon
	} from '@nais/ds-svelte-community/icons';

	const { icon } = $props();

	const getComponent = (icon: string) => {
		switch (icon.toLocaleLowerCase()) {
			case 'team overview':
			case 'oversikt':
				return HexagonGridIcon;
			case 'medlemmer':
			case 'medlemskap':
			case 'groups':
			case 'grupper':
				return PersonGroupIcon;
			case 'settings':
			case 'activity log':
			case 'aktivitetslogg':
				return ShieldLockIcon;
			case 'datadeling':
			case 'datatilgang':
				return FloppydiskIcon;
			default:
				return QuestionmarkIcon;
		}
	};

	const Component = $derived(getComponent(icon));
</script>

<Component />
