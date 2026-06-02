import { env } from '$env/dynamic/private';
import { ServerGetUserStore } from '$houdini';
import { createFeedbackMessage } from './message';
import { json } from '@sveltejs/kit';

const webhookUrl = env.SLACK_WEBHOOK_URL || '';

export async function POST(event) {
	const { request } = event;

	const q = new ServerGetUserStore();
	const { data } = await q.fetch({ event });
	if (data?.me.__typename !== 'User') {
		return json({ error: 'Not authenticated' }, { status: 401 });
	}
	if (!data?.me.email) {
		return json({ error: 'Not authenticated' }, { status: 401 });
	}

	const email = data.me.email;

	const body = await request.json();
	const { anonymous, feedback, path, type } = body;

	try {
		const result = await fetch(webhookUrl, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				feedback: createFeedbackMessage(anonymous, email, feedback, path, type)
			})
		});

		if (result.ok) {
			return json({ message: 'Feedback sent successfully!' });
		} else {
			return json({ error: 'Failed to send feedback' }, { status: 500 });
		}
	} catch (error) {
		return json({ error: 'An error occurred - ' + error }, { status: 500 });
	}
}
