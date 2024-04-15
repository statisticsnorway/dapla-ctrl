// https://vitejs.dev/config/server-options.html
// https://github.com/garronej/vite-envs

import { DropdownItems } from '../@types/pageTypes'
import { JobResponse } from '../services/teamDetail'
import { Team } from '../services/sharedBucketDetail'

export const DAPLA_TEAM_API_URL = `/api`

export const getGroupType = (teamName: string, groupName: string): string => {
  if (teamName === undefined || groupName === undefined || !teamName.length || !groupName.length) return ''
  return groupName.slice(teamName.length + 1)
}

// TODO: Replaced by the getTeamFromGroup function; consider removing
// export const stripSuffixes = (inputString: string) => {
//   // Regular expression to match the specified suffixes

//   // TODO: Get suffixes from function parameters (which is passed from query response), this because we dont know
//   // all the custom groups, and they can be created at any time
//   // this may require a re-write of how we fetch and aggregate data in SharedBucketDetail.tsx
//   const suffixesPattern = /-(data-admins|managers|developers|consumers|support|editor|admin)$/

//   // Replace matched suffix with an empty string
//   return inputString.replace(suffixesPattern, '')
// }

export const getTeamFromGroup = (allTeams: Team[], groupName: string): string => {
  if (allTeams?.length) {
    const teamName = allTeams.filter(({ uniform_name }) => {
      if (groupName.includes(uniform_name)) return uniform_name
    })
    return teamName[0].uniform_name
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

export const removeDuplicateDropdownItems = (items: DropdownItems[]) => {
  return items.reduce((acc: DropdownItems[], dropdownItem: DropdownItems) => {
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
