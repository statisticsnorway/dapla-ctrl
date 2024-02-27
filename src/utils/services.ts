export const fetchAPIData = async (url: string, accessToken: string): Promise<any> => {
  const response = await fetch(url, {
    method: 'GET',
    headers: {
      accept: '*/*',
      Authorization: `Bearer ${accessToken}`,
    },
  })

  if (!response.ok) {
    const errorMessage = (await response.text()) || 'An error occurred'
    const { detail, status } = JSON.parse(errorMessage)
    throw new ApiError(status, detail)
  }

  return await response.json()
}

export class ApiError extends Error {
  public code: number

  constructor(code: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.code = code
  }
}
