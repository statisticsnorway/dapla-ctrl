import { Data } from 'effect'

// Represents an error when converting between representations of some data
export class ConversionError extends Data.TaggedError('ConversionError')<{ message: string }> {
  override get message() {
    return this.message_
  }

  private readonly message_: string

  constructor(message: string) {
    super({ message })
    this.message_ = message
  }
}

// Represents the error where the user isn't logged in
export class UserNotLoggedIn extends Data.TaggedError('UserNotLoggedIn')<{ message: string }> {
  override get message() {
    return this.message_
  }

  private readonly message_: string

  constructor(message: string) {
    super({ message })
    this.message_ = message
  }
}
