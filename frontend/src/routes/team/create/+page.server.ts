import { env } from '$env/dynamic/private';
import type { Actions } from './$types';

const webhookUrl = env.SLACK_CREATE_TEAM_WEBHOOK_URL || '';

export const actions = {
	default: async (event) => {
		const data = await event.request.formData();
		const displayName = (data.get('displayname') as string) || '';
		const input = {
			slug: (data.get('name') as string) || '',
			displayName,
			sectionCode: (data.get('section') as string) || '',
			isManaged: (data.get('isManaged') as string) || 'true',
			createdBy: (data.get('createdBy') as string) || ''
		};

		try {
			const result = await fetch(webhookUrl, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(input)
			});

			if (result.ok) {
				return { success: true, input, message: 'Forespørsel om å opprette team er sendt!' };
			} else {
				return {
					success: false,
					input,
					errors: [{ message: 'En feil oppsto under opprettelse av team' }]
				};
			}
		} catch (error) {
			return { success: false, input, errors: [{ message: 'En feil oppsto - ' + error }] };
		}
	}
} satisfies Actions;
