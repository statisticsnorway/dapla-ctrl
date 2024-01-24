export async function getRequest<T>(path: string, token: string | null): Promise<T> {
	if (token === null) {
		console.error("Token was null")
	}

    return fetch(path, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        }
    })
}

export async function postRequest<T>(path: string, token: string | null, body: string | null): Promise<T> {
	if (token === null) {
		console.error("Token was null")
	}

	return fetch(path, {
		method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token || accessToken}`
        },
		body: body,
	})
}
