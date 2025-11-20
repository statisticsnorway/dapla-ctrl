<script lang="ts">
	import {
		CogIcon,
		HouseIcon,
		PersonGroupIcon,
		QuestionmarkIcon,
		ShieldLockIcon,
		TrashIcon
	} from '@nais/ds-svelte-community/icons';

	const { icon } = $props();

	const getComponent = (icon: string) => {
		switch (icon.toLocaleLowerCase()) {
			case 'team overview':
				return HouseIcon;
			case 'groups':
				return PersonGroupIcon;
			case 'settings':
				return CogIcon;
			case 'activity log':
				return ShieldLockIcon;
			case 'delete':
				return TrashIcon;
			default:
				return QuestionmarkIcon;
		}
	};

	const Component = getComponent(icon);
</script>

<Component />
