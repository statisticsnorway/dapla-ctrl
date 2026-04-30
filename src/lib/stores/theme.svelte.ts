import { browser } from '$app/environment';

export type Themes = 'dark' | 'light';

export const persistTheme = (theme: Themes) => {
	if (browser) {
		const formData = new FormData();
		formData.append('theme', theme);
		fetch('/api/theme', {
			method: 'POST',
			body: formData
		});
	}
};
