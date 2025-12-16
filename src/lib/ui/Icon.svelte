<script lang="ts">
	import {
		BellIcon,
		BranchingIcon,
		BriefcaseClockIcon,
		BucketIcon,
		CogIcon,
		DatabaseIcon,
		Density3Icon,
		FileTextIcon,
		HexagonGridIcon,
		ImageIcon,
		LineGraphStackedIcon,
		LinkIcon,
		PackageIcon,
		PadlockLockedIcon,
		PersonGroupIcon,
		PersonIcon,
		QuestionmarkIcon,
		RocketIcon,
		ShieldLockIcon,
		TrashIcon,
		VirusIcon,
		VitalsIcon,
		WalletIcon
	} from '@nais/ds-svelte-community/icons';

	const { icon } = $props();

	const getComponent = (icon: string) => {
		switch (icon.toLocaleLowerCase()) {
			case 'team overview':
			case 'teamoversikt':
				return HexagonGridIcon;
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
			case 'buckets':
			case 'bucket':
				return BucketIcon;
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
			case 'groups':
			case 'grupper':
				return PersonGroupIcon;
			case 'repositories':
				return BranchingIcon;
			case 'settings':
			case 'instillinger':
				return CogIcon;
			case 'activity log':
			case 'aktivitetslogg':
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
			case 'ingresses':
				return LinkIcon;
			case 'alerts':
				return BellIcon;
			case 'issues':
				return VitalsIcon;
			case 'brukeroversikt':
				return PersonIcon;
			default:
				return QuestionmarkIcon;
		}
	};

	const Component = getComponent(icon);
</script>

<Component />
