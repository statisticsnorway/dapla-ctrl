<script lang="ts">
	import {
		HexagonGridIcon,
		PersonGroupIcon,
		PersonIcon,
		QuestionmarkIcon,
		ShieldLockIcon
	} from '@nais/ds-svelte-community/icons';

	const { icon } = $props();

	const getComponent = (icon: string) => {
		switch (icon.toLocaleLowerCase()) {
			case 'team overview':
			case 'oversikt':
				return HexagonGridIcon;
			case 'members':
			case 'groups':
			case 'grupper':
				return PersonGroupIcon;
			case 'settings':
			case 'activity log':
			case 'aktivitetslogg':
				return ShieldLockIcon;
			case 'brukeroversikt':
				return PersonIcon;
			default:
				return QuestionmarkIcon;
		}
	};

	const Component = getComponent(icon);
</script>

<Component />
