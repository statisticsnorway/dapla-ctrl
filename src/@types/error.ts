export interface ErrorResponse {
  error: Error
}

export interface Error {
  code: string
  message: string
}
