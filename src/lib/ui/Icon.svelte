<script lang="ts">
	import {
		CogIcon,
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
			case 'teamoversikt':
				return HexagonGridIcon;
			case 'members':
			case 'groups':
			case 'grupper':
				return PersonGroupIcon;
			case 'settings':
			case 'instillinger':
				return CogIcon;
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
