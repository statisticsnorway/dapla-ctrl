import { menuItems } from './menuItems';

describe('menuItems', () => {
	describe('team menu', () => {
		test('full', () => {
			expect(
				menuItems({
					path: '/team/devteam',
					isManaged: true,
					member: true,
					isAdmin: false
				})
			).toEqual([
				[{ label: 'Oversikt', href: '/team/devteam', active: true }],
				[
					{ label: 'Medlemmer', href: '/team/devteam/members' },
					{ label: 'Datadeling', href: '/team/devteam/shared-data' },
					{ label: 'Dapla Lab', href: '/team/devteam/launch-lab' },
					{ label: 'Aktivitetslogg', href: '/team/devteam/activity-log' },
					{ label: 'Innstillinger', href: '/team/devteam/settings' }
				]
			]);
		});
	});
});
