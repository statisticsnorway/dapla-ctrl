<script module>
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Menu from './Menu.svelte';

	const { Story } = defineMeta({
		component: Menu,
		tags: ['autodocs']
	});
</script>

<script lang="ts">
</script>

<Story name="Team menu">
	<Menu items={[[{ label: 'Overview', href: '/team/devteam' }], [], [], [], []]} />
</Story>

<Story name="Application menu">
	<Menu
		items={[
			[
				{ label: 'Overview', href: '/team/devteam/dev/app/app-w-all-storage' },
				{ label: 'Status', href: '/team/devteam/dev/app/app-w-all-storage/status' }
			],
			[
				{
					label: 'Vulnerability Report',
					href: '/team/devteam/dev/app/app-w-all-storage/vulnerability-report'
				},
				{ label: 'Deployments', href: '/team/devteam/dev/app/app-w-all-storage/deploys' },
				{ label: 'Cost', href: '/team/devteam/dev/app/app-w-all-storage/cost' },
				{
					label: 'Utilization',
					href: '/team/devteam/dev/app/app-w-all-storage/utilization',
					active: true
				},
				{ label: 'Logs', href: '/team/devteam/dev/app/app-w-all-storage/logs' }
			],
			[{ label: 'Manifest', href: '/team/devteam/dev/app/app-w-all-storage/manifest' }],
			[{ label: 'Delete', href: '/team/devteam/dev/app/app-w-all-storage/delete' }]
		]}
	/>
</Story>
