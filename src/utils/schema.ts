import { Schema as S } from 'effect'
import { Schema, PropertySignature } from 'effect/Schema'

/**
 * Utility function to make tagging fields with different encoding key names less verbose.
 * @param {string}                    key - The name of the encoded key.
 * @param {Schema.Schema<A, I, R>} schema - The schema to be modified with the encoding.
 * @returns {Schema.PropertySignature<':', A, string, ':', I, false, R>}
 * */
export const withKeyEncoding = <A, I, R>(
  key: string,
  schema: Schema<A, I, R>
): PropertySignature<':', A, string, ':', I, false, R> => S.fromKey(key)(S.propertySignature(schema))
