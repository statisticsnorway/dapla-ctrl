import { Console, Effect, Logger } from 'effect'
import { HttpClient, HttpClientRequest, FetchHttpClient } from '@effect/platform'
import * as FiberId from 'effect/FiberId'

const logger = Logger.make((options) => {
  const data = {
    fiberId: FiberId.threadName(options.fiberId),
    timestamp: options.date,
    logLevel: options.logLevel.label,
    message: options.message,
  }
  const logMsg = JSON.stringify(data, null, 2)

  const logRequest = Effect.gen(function* () {
    const client = yield* HttpClient.HttpClient
    return yield* HttpClientRequest.post('/log').pipe(HttpClientRequest.bodyText(logMsg, 'utf-8'), client.execute)
  }).pipe(Effect.scoped, Effect.provide(FetchHttpClient.layer))

  return Console.log(logMsg).pipe(Effect.zipRight(logRequest), Effect.runPromise)
})

export const customLogger = Logger.replace(Logger.defaultLogger, logger)
