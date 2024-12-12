import { Effect } from 'effect'
import { ParseResult, Schema } from 'effect'
// import * as Http from '@effect/platform/HttpClient'
// import { HttpClientError } from '@effect/platform/Http/ClientError'
// import * as ClientResponse from '@effect/platform/HttpClientResponse'
// import { BodyError } from '@effect/platform/Http/Body'

import { FetchHttpClient, HttpClientRequest, HttpClientResponse } from '@effect/platform'
import { HttpClient } from '@effect/platform/HttpClient'
import { HttpClientError } from '@effect/platform/HttpClientError'
import { HttpBodyError } from '@effect/platform/HttpBody'

import { DAPLA_TEAM_API_URL } from '../utils/utils'
import { withKeyEncoding } from '../utils/schema'
import { customLogger } from '../utils/logger'

const CREATE_TEAM_URL = `${DAPLA_TEAM_API_URL}/teams/create`

const FeatureSchema = Schema.Literal('kildomaten', 'daplabuckets', 'transferservice')

const AutonomyLevelSchema = Schema.Literal('managed', 'semi-managed', 'self-managed')

const CreateTeamRequestSchema = Schema.Struct({
  uniformTeamName: withKeyEncoding('uniform_team_name', Schema.String),
  teamDisplayName: withKeyEncoding('team_display_name', Schema.String),
  sectionCode: withKeyEncoding('section_code', Schema.String),
  additionalInformation: withKeyEncoding('additional_information', Schema.String),
  autonomyLevel: withKeyEncoding('autonomy_level', AutonomyLevelSchema),
  features: Schema.Array(FeatureSchema),
})

export type AutonomyLevel = Schema.Schema.Type<typeof AutonomyLevelSchema>
export type Feature = Schema.Schema.Type<typeof FeatureSchema>
export type CreateTeamRequest = Schema.Schema.Type<typeof CreateTeamRequestSchema>

const CreateTeamResponseSchema = Schema.Struct({
  uniformTeamName: withKeyEncoding('uniform_team_name', Schema.String),
  kubenPullRequestId: withKeyEncoding('kuben_pull_request_id', Schema.Number),
  kubenPullRequestUrl: withKeyEncoding('kuben_pull_request_url', Schema.String),
})

export type CreateTeamResponse = Schema.Schema.Type<typeof CreateTeamResponseSchema>

export const isAuthorizedToCreateTeam = (isDaplaAdmin: boolean, userJobTitle: string) =>
  isDaplaAdmin || ['seksjonssjef', 'forskningsleder'].includes(userJobTitle.toLowerCase())

export const createTeam = (
  createTeamRequest: CreateTeamRequest
): Effect.Effect<CreateTeamResponse, HttpBodyError | HttpClientError | ParseResult.ParseError> =>
  Effect.zipRight(
    Effect.logInfo('CreateTeamRequest', createTeamRequest).pipe(Effect.provide(customLogger)),
    HttpClient.pipe(
      Effect.flatMap((client) =>
        HttpClientRequest.post(new URL(CREATE_TEAM_URL, window.location.origin)).pipe(
          HttpClientRequest.schemaBodyJson(CreateTeamRequestSchema)(createTeamRequest),
          Effect.flatMap(client.execute)
        )
      ),
      Effect.flatMap(HttpClientResponse.schemaBodyJson(CreateTeamResponseSchema)),
      Effect.scoped,
      Effect.provide(FetchHttpClient.layer)
    )
    //Http.request
    //  .post(new URL(CREATE_TEAM_URL, window.location.origin))
    //  .pipe(
    //    Http.request.schemaBody(CreateTeamRequestSchema)(createTeamRequest),
    //    Effect.flatMap(Http.client.fetchOk),
    //    ClientResponse.schemaBodyJsonScoped(CreateTeamResponse)
    //  )
  )
