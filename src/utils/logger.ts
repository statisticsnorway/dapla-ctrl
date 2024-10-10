import { Console, Effect, Logger } from 'effect'
import * as Http from '@effect/platform/HttpClient'
import * as FiberId from 'effect/FiberId'

const logger = Logger.make((options) => {
  const data = {
    fiberId: FiberId.threadName(options.fiberId),
    timestamp: options.date,
    logLevel: options.logLevel.label,
    message: options.message,
  }
  const logMsg = JSON.stringify(data, null, 2)
  return Console.log(logMsg).pipe(
    Effect.zipRight(Http.request.post('/log').pipe(Http.request.textBody(logMsg), Http.client.fetchOk, Effect.scoped)),
    Effect.runPromise
  )
})

export const customLogger = Logger.replace(Logger.defaultLogger, logger)
