export function getGroupType(groupName: string) {
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
