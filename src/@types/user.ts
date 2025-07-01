import { Schema } from 'effect'
import { withKeyEncoding } from '../utils/schema'

export interface User {
  principal_name: string
  azure_ad_id: string
  display_name: string
  first_name: string
  last_name: string
  email: string
  division_name?: string
  division_code?: number
  section_name?: string
  section_code?: number
  section_manager?: User
  photo?: string
}

export const UserData = Schema.Struct({
  principalName: withKeyEncoding('principal_name', Schema.String),
  azureAdId: withKeyEncoding('azure_ad_id', Schema.String),
  displayName: withKeyEncoding('display_name', Schema.String),
  firstName: withKeyEncoding('first_name', Schema.String),
  lastName: withKeyEncoding('last_name', Schema.String),
  email: Schema.String,
  phone: Schema.optional(Schema.String),
  jobTitle: withKeyEncoding('job_title', Schema.String),
  divisionName: withKeyEncoding('division_name', Schema.String),
  divisionCode: withKeyEncoding('division_code', Schema.NumberFromString),
  sectionName: withKeyEncoding('section_name', Schema.String),
  sectionCode: withKeyEncoding('section_code', Schema.NumberFromString),
})

export type UserData = typeof UserData.Type

export const UserPhoto = Schema.Struct({
  photo: Schema.String,
})

export type UserPhoto = typeof UserPhoto.Type

// This type is used to represent a user logged in to dapla-ctrl
export const UserProfile = Schema.extend(UserData, UserPhoto)

export type UserProfile = typeof UserProfile.Type
