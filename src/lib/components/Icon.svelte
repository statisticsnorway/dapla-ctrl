<script lang="ts">
	import {
		BellIcon,
		BranchingIcon,
		BriefcaseClockIcon,
		CogIcon,
		DatabaseIcon,
		Density3Icon,
		FileTextIcon,
		HouseIcon,
		ImageIcon,
		LineGraphStackedIcon,
		PackageIcon,
		PadlockLockedIcon,
		PersonGroupIcon,
		QuestionmarkIcon,
		RocketIcon,
		ShieldLockIcon,
		TrashIcon,
		VirusIcon,
		WalletIcon
	} from '@nais/ds-svelte-community/icons';

	const { icon } = $props();

	const getComponent = (icon: string) => {
		switch (icon.toLocaleLowerCase()) {
			case 'team overview':
				return HouseIcon;
			case 'app overview':
			case 'applications':
				return PackageIcon;
			case 'job overview':
			case 'jobs':
				return BriefcaseClockIcon;
			case 'secrets':
				return PadlockLockedIcon;
			case 'sqlinstance':
			case 'postgres':
				return DatabaseIcon;
			case 'deployments':
				return RocketIcon;
			case 'cost':
				return WalletIcon;
			case 'utilization':
				return LineGraphStackedIcon;
			case 'vulnerabilities':
			case 'vulnerability report':
				return VirusIcon;
			case 'members':
				return PersonGroupIcon;
			case 'repositories':
				return BranchingIcon;
			case 'settings':
				return CogIcon;
			case 'activity log':
				return ShieldLockIcon;
			case 'status':
				return BellIcon;
			case 'image':
				return ImageIcon;
			case 'logs':
				return Density3Icon;
			case 'manifest':
				return FileTextIcon;
			case 'delete':
				return TrashIcon;
			default:
				return QuestionmarkIcon;
		}
	};

	const Component = getComponent(icon);
</script>

<Component />
