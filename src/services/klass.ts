import { Array as A, Effect, Either, Option as O, pipe } from 'effect'
import { Schema } from 'effect'
import { ParseError } from 'effect/ParseResult'
import { FetchHttpClient, HttpClientRequest } from '@effect/platform'
import { HttpClient } from '@effect/platform/HttpClient'
import { HttpClientError } from '@effect/platform/HttpClientError'

const KLASS_URL = '/klass'

const OrgUnitVersionsSchema = Schema.NonEmptyArray(
  Schema.Struct({
    name: Schema.String,
    validFrom: Schema.String,
    validTo: Schema.optional(Schema.String),
    _links: Schema.Struct({ self: Schema.Struct({ href: Schema.String }) }),
  })
)

export type OrgUnitVersions = typeof OrgUnitVersionsSchema.Type

const fetchOrgUnitVersions = (): Effect.Effect<OrgUnitVersions, HttpClientError | ParseError> =>
  HttpClient.pipe(
    Effect.flatMap((client) =>
      HttpClientRequest.get(new URL(`${KLASS_URL}/classifications/83`, window.location.origin)).pipe(
        HttpClientRequest.appendUrlParams({ language: 'en', includeFuture: true }),
        client.execute,
        Effect.flatMap((res) => res.json)
      )
    ),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    Effect.flatMap((jsonResponse: any) =>
      Either.match(Schema.decodeEither(OrgUnitVersionsSchema)(jsonResponse.versions), {
        onLeft: (error) => Effect.fail(error),
        onRight: (v: OrgUnitVersions) => Effect.succeed(v),
      })
    ),
    Effect.scoped,
    Effect.provide(FetchHttpClient.layer)
  )

//  Http.request.get(new URL(`${KLASS_URL}/classifications/83`, window.location.origin)).pipe(
//    Http.request.appendUrlParam('language', 'en'),
//    Http.request.appendUrlParam('includeFuture', 'true'),
//    Http.client.fetchOk,
//    Http.response.json,
//    // eslint-disable-next-line @typescript-eslint/no-explicit-any
//    Effect.flatMap((jsonResponse: any) =>
//      Either.match(Schema.decodeEither(OrgUnitVersionsSchema)(jsonResponse.versions), {
//        onLeft: (error) => Effect.fail(error),
//        onRight: (v: OrgUnitVersions) => Effect.succeed(v),
//      })
//    ),
//    Effect.scoped
//  )

const SSBSectionSchema = Schema.Struct({
  code: Schema.NumberFromString,
  parentCode: Schema.String,
  level: Schema.NumberFromString,
  name: Schema.String,
})

export const SSBSectionsSchema = Schema.Array(SSBSectionSchema)

export type SSBSection = typeof SSBSectionSchema.Type

export type SSBSections = typeof SSBSectionsSchema.Type

// TODO: Remove this if we ever figure out how to call klass without proxying the API
// due to CORS errors.
const mapToProxyUrl = (str: string) => pipe(str.match(/\/versions\/\d+/g), O.fromNullable, O.flatMap(A.head))

/**
 * Filter out departments from list of sections.
 */
const removeSSBDepartments = (sections: SSBSections): SSBSections =>
  A.filter(sections, (section) => section.parentCode !== '')

const fetchSSBSectionsAndDepartments = (versionUrl: string): Effect.Effect<SSBSections, HttpClientError | ParseError> =>
  HttpClient.pipe(
    Effect.flatMap((client) => client.get(`/klass${O.getOrThrow(mapToProxyUrl(versionUrl))}`)),
    Effect.flatMap((res) => res.json),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    Effect.flatMap((jsonResponse: any) => {
      return Either.match(Schema.decodeEither(SSBSectionsSchema)(jsonResponse.classificationItems), {
        onLeft: (error) => Effect.fail(error),
        onRight: (v: SSBSections) => Effect.succeed(v),
      })
    })
  ).pipe(Effect.scoped, Effect.provide(FetchHttpClient.layer))

export const fetchSSBSectionInformation = (): Effect.Effect<SSBSections, HttpClientError | ParseError, never> =>
  Effect.gen(function* () {
    const orgUnitVersions: OrgUnitVersions = yield* fetchOrgUnitVersions()
    const url = pipe(orgUnitVersions, A.lastNonEmpty, (orgUnitVersion) => orgUnitVersion._links.self.href)
    const sections = yield* fetchSSBSectionsAndDepartments(url)
    return removeSSBDepartments(sections)
  })
