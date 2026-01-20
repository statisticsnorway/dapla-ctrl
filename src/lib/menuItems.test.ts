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
					{ label: 'Grupper', href: '/team/devteam/groups' },
					{ label: 'Aktivitetslogg', href: '/team/devteam/activity-log' }
				]
			]);
		});
	});
});
