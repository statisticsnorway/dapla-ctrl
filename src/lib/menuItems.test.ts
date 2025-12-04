import { menuItems } from './menuItems';

const features = {
	valkey: { enabled: true },
	openSearch: { enabled: true },
	kafka: { enabled: true },
	unleash: { enabled: true }
};

describe('menuItems', () => {
	describe('team menu', () => {
		test('full', () => {
			expect(
				menuItems({
					path: '/team/devteam',
					features,
					member: true,
					isAdmin: false
				})
			).toEqual([
				[{ label: 'Team Overview', href: '/team/devteam', active: true }],
				[
					{ label: 'Groups', href: '/team/devteam/groups' },
					{ label: 'Settings', href: '/team/devteam/settings' },
					{ label: 'Activity Log', href: '/team/devteam/activity-log' }
				]
			]);
		});

		test('when not member', () => {
			expect(
				menuItems({
					path: '/team/tbd/jobs',
					features,
					member: false,
					isAdmin: false
				})
					.flatMap((g) => g)
					.find((i) => ['Settings'].includes(i.label))
			).toBeUndefined();
		});

		test('inventory', () => {
			const res = menuItems({
				path: '/team/tbd/jobs',
				features,
				member: true,
				isAdmin: false
			});

			expect(
				res
					.flatMap((g) => g)
					.filter((i) => i.count)
					.map((i) => ({ label: i.label, count: i.count }))
			).toEqual([]);
		});
		test('show settings when admin', () => {
			expect(
				menuItems({
					path: '/team/nais',
					features,
					member: false,
					isAdmin: true
				})
					.flatMap((g) => g)
					.find((i) => ['Settings'].includes(i.label))
			).toBeDefined();
		});
	});
});
