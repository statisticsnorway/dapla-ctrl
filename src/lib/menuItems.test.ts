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
				[{ label: 'Teamoversikt', href: '/team/devteam', active: true }],
				[
					{ label: 'Grupper', href: '/team/devteam/groups' },
					{ label: 'Instillinger', href: '/team/devteam/settings' },
					{ label: 'Aktivitetslogg', href: '/team/devteam/activity-log' }
				]
			]);
		});

		test('when not member', () => {
			expect(
				menuItems({
					path: '/team/tbd/settings',
					member: false,
					isAdmin: false
				})
					.flatMap((g) => g)
					.find((i) => ['Instillinger'].includes(i.label))
			).toBeUndefined();
		});

		test('show settings when admin', () => {
			expect(
				menuItems({
					path: '/team/nais',
					member: false,
					isAdmin: true
				})
					.flatMap((g) => g)
					.find((i) => ['Instillinger'].includes(i.label))
			).toBeDefined();
		});
	});
});
