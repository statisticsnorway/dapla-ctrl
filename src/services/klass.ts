import { Array as A, Effect, Either, Option as O, pipe } from 'effect'
import { Schema } from '@effect/schema'
import { ParseError } from '@effect/schema/ParseResult'
import * as Http from '@effect/platform/HttpClient'
import { HttpClientError } from '@effect/platform/Http/ClientError'

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

const fetchOrgUnitVersions = (): Effect.Effect<OrgUnitVersions, HttpClientError | ParseError, never> =>
    Http.request.get(new URL(`${KLASS_URL}/classifications/83`, window.location.origin)).pipe(
        Http.request.appendUrlParam('language', 'en'),
        Http.request.appendUrlParam('includeFuture', 'true'),
        Http.client.fetchOk,
        Http.response.json,
        Effect.flatMap((jsonResponse: any) =>
            Either.match(Schema.decodeEither(OrgUnitVersionsSchema)(jsonResponse.versions), {
                onLeft: (error) => Effect.fail(error),
                onRight: (v: OrgUnitVersions) => Effect.succeed(v),
            })
        ),
        Effect.scoped
    )

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

const fetchSSBSectionsAndDepartments = (
    versionUrl: string
): Effect.Effect<SSBSections, HttpClientError | ParseError, never> =>
    Http.request.get(`/klass${O.getOrThrow(mapToProxyUrl(versionUrl))}`).pipe(
        Http.client.fetchOk,
        Http.response.json,
        Effect.flatMap((jsonResponse: any) => {
            return Either.match(Schema.decodeEither(SSBSectionsSchema)(jsonResponse.classificationItems), {
                onLeft: (error) => Effect.fail(error),
                onRight: (v: SSBSections) => Effect.succeed(v),
            })
        })
    )

export const fetchSSBSectionInformation = (): Effect.Effect<SSBSections, HttpClientError | ParseError, never> =>
    Effect.gen(function* (_) {
        const orgUnitVersions: OrgUnitVersions = yield* fetchOrgUnitVersions()
        const url = pipe(orgUnitVersions, A.lastNonEmpty, (orgUnitVersion) => orgUnitVersion._links.self.href)
        const sections = yield* fetchSSBSectionsAndDepartments(url)
        return removeSSBDepartments(sections)
    })
