import { Console, Effect, Logger } from 'effect'
import * as Http from '@effect/platform/HttpClient'

const logger = Logger.make(({ logLevel, message }) => {
  const logMsg = `[${logLevel.label}] ${message}`
  Console.log(logMsg).pipe(
    Effect.zipRight(Http.request.post('/log').pipe(Http.request.textBody(logMsg), Http.client.fetchOk, Effect.scoped)),
    Effect.runPromise
  )
})

export const customLogger = Logger.replace(Logger.defaultLogger, logger)
