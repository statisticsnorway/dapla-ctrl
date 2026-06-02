import type { FeedbackType } from '$lib/feedback/types';

export function createFeedbackMessage(
	anonymous: boolean,
	email: string,
	feedback: string,
	path: string,
	type: FeedbackType
): string {
	let headerText: string;

	if (anonymous) {
		email = 'Anonymous';
	}

	switch (type) {
		case 'KUDOS':
			headerText = ':sparkles:\nKudos';
			break;
		case 'BUG':
			headerText = ':bug:\nBug report';
			break;
		case 'CHANGE_REQUEST':
			headerText = ':bulb:\nChange request';
			break;
		case 'OTHER':
			headerText = ':speech_balloon:\nOther feedback';
			break;
		case 'QUESTION':
			headerText = ':question:\nQuestion';
			break;
		default:
			headerText = type;
	}

	const details = [`From: ${email}`, `Path: ${path}`];

	return `${headerText}
${details.join('\n')}
Feedback:
${feedback}`;
}
