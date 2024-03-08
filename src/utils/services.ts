// eslint-disable-next-line
export const fetchAPIData = async (url: string): Promise<any> => {
  const response = await fetch(url)

  if (!response.ok) {
    const errorMessage = (await response.text()) || 'An error occurred'
    const { detail, status } = JSON.parse(errorMessage)
    throw new ApiError(status, detail)
  }

  return response.json()
}

interface TokenData {
  name: string
  given_name: string
  family_name: string
  email: string
}

export const fetchUserInformationFromAuthToken = async (): Promise<TokenData> => {
  const response = await fetch('/localApi/fetch-token')

  const tokenData = await response.json()
  const jwt = JSON.parse(atob(tokenData.token.split('.')[1]))
  return { ...jwt } as TokenData
}

export class ApiError extends Error {
  public code: number

  constructor(code: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.code = code
  }
}
