// https://vitejs.dev/config/server-options.html
// https://github.com/garronej/vite-envs

export const DAPLA_TEAM_API_URL = `/api`

export const getGroupType = (groupName: string) => {
  const match = groupName.match(/(managers|developers|data-admins|support|consumers)$/)
  const role = match ? match[0] : null
  switch (role) {
    case 'managers':
      return 'Manager'
    case 'developers':
      return 'Developers'
    case 'data-admins':
      return 'Data-admins'
    case 'support':
      return 'Support'
    case 'consumers':
      return 'Consumers'
    default:
      return groupName
  }
}

export const stripSuffixes = (inputString: string) => {
  // Regular expression to match the specified suffixes
  const suffixesPattern = /-(data-admins|managers|developers|consumers|support)$/

  // Replace matched suffix with an empty string
  return inputString.replace(suffixesPattern, '')
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
