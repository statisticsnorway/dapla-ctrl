import type { ClientPlugin } from '$houdini';
import { writable } from 'svelte/store';
import { redirect } from '@sveltejs/kit';
import { page } from '$app/state';

export const isAuthenticated = writable<boolean>(true);

export const isUnauthenticated = (errors: { message: string }[] | null) => {
	const unauthenticatedError = 'Unauthorized';
	if (
		errors &&
		errors.length > 0 &&
		errors.filter((error) => error.message === unauthenticatedError).length > 0
	) {
		return true;
	}
	return false;
};

export const handleMissingLogin = (...ignoredNames: string[]): ClientPlugin => {
	return () => {
		return {
			afterNetwork(ctx, { value, resolve }) {
				if (!ignoredNames.includes(ctx.name) && isUnauthenticated(value.errors)) {
					isAuthenticated.set(false);
					const redirectPath = (url: URL) => {
						return encodeURIComponent(url.pathname + url.search + url.hash);
					};
					const oauth2LoginPath = '/oauth2/login?redirect_uri=' + redirectPath(page.url);
					redirect(302, oauth2LoginPath);
				} else if (ctx.name == 'UserInfo' && value.data) {
					if (value.data.me) {
						isAuthenticated.set(true);
					}
				}
				return resolve(ctx);
			}
		};
	};
};
