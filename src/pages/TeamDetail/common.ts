import { getGroupType } from '../../utils/utils.ts'
import { Team, Group } from '../../services/teamDetail.ts'
import { DropdownItem } from '../../@types/pageTypes'

// The only valid groups for MANAGED dapla teams
export const standardGroups = ['developers', 'managers', 'data-admins']

/**
 * Display groups for use in the Dropdown component. Disable the group
 * if it's non-standard.
 *
 */
export const displayGroupItem =
  (teamDetailData: Team) =>
  ({ uniform_name }: Group): DropdownItem => {
    const groupType = getGroupType(teamDetailData.uniform_name, uniform_name)
    return {
      id: uniform_name,
      title: groupType,
      disabled: !standardGroups.includes(groupType),
    }
  }
