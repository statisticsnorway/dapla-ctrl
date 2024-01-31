export const validateKeycloakToken = (token?: string): Promise<boolean> => {
  const accessToken = localStorage.getItem('access_token')

  return fetch('/api/verify-token', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token || accessToken}`,
    },
  })
    .then((response) => {
      if (!response.ok) {
        console.error('Token verification failed with status:', response.status)
        return false
      }
      return response.status === 200
    })
    .catch((error) => {
      console.error('Error during token validation:', error)
      return false
    })
}
