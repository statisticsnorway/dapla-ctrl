// https://vitejs.dev/config/server-options.html
// https://github.com/garronej/vite-envs

import { DropdownItem } from '../@types/pageTypes'
import { JobResponse } from '../services/teamDetail'
import { Team } from '../services/sharedBucketDetail'
import { Option as O } from 'effect'
import { LazyArg, dual } from 'effect/Function'

export const DAPLA_TEAM_API_URL = `/api`

/**
 * Eliminator for the Option type. Given a mapping function and a default value
 * applies the mapping function and returns the result if Option type contains a value,
 * otherwise returns the default value.
 *
 * @example
 * import { Option as O } from 'effect'
 * const result = option(O.none(), -1, (x) => x + 5)
 *
 * assert.deepStrictEqual(result, -1)
 *
 * @example
 * import { Option as O } from 'effect'
 * const result = option(O.some(10), -1, (x) => x + 5)
 *
 * assert.deepStrictEqual(result, 15)
 *
 */
export const option: {
  <B, A>(z: LazyArg<B>, f: (a: A) => B): (value: O.Option<A>) => B
  <A, B>(value: O.Option<A>, z: LazyArg<B>, f: (a: A) => B): B
} = dual(3, <A, B>(value: O.Option<A>, z: LazyArg<B>, f: (a: A) => B): B => value.pipe(O.map(f), O.getOrElse(z)))

export const getGroupType = (teamName: string, groupName: string): string => {
  if (teamName === undefined || groupName === undefined || !teamName.length || !groupName.length) return ''
  return groupName.slice(teamName.length + 1)
}

// Returns the closest match of a team name from a group name
export const getTeamFromGroup = (allTeams: Team[], groupName: string): string => {
  if (allTeams?.length) {
    const teamName = allTeams
      .filter(({ uniform_name }) => groupName.includes(uniform_name))
      .sort((a, b) => b.uniform_name.length - a.uniform_name.length)
    return teamName.length ? teamName[0].uniform_name : ''
  }
  return ''
}

export const formatDisplayName = (displayName: string) => {
  return displayName.split(', ').reverse().join(' ')
}

// eslint-disable-next-line
export const flattenEmbedded = (json: any): any => {
  if (json._embedded) {
    for (const prop in json._embedded) {
      json[prop] = json._embedded[prop]
    }
    delete json._embedded
  }

  for (const prop in json) {
    if (typeof json[prop] === 'object') {
      json[prop] = flattenEmbedded(json[prop])
    }
  }

  return json
}

export const removeDuplicateDropdownItems = (items: DropdownItem[]) => {
  return items.reduce((acc: DropdownItem[], dropdownItem: DropdownItem) => {
    const ids = acc.map((obj) => obj.id)
    if (!ids.includes(dropdownItem.id)) {
      acc.push(dropdownItem)
    }
    return acc
  }, [])
}

export const getErrorList = (response: JobResponse[]) => {
  return response
    .map(({ status, detail }) => {
      if ((detail && status === 'ERROR') || (detail && status === 'IGNORED')) {
        return detail
      }
      return ''
    })
    .filter((str) => str !== '')
}
