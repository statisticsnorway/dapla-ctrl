import { menuItems } from './menuItems';

describe('menuItems', () => {
	describe('team menu', () => {
		test('full', () => {
			expect(
				menuItems({
					path: '/team/devteam',
					member: true,
					isAdmin: false
				})
			).toEqual([
				[{ label: 'Oversikt', href: '/team/devteam', active: true }],
				[
					{ label: 'Medlemmer', href: '/team/devteam/members' },
					{ label: 'Datadeling', href: '/team/devteam/shared-data' },
					{ label: 'Aktivitetslogg', href: '/team/devteam/activity-log' }
				]
			]);
		});
	});
});
