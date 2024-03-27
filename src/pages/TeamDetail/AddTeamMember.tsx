import styles from './teamDetail.module.scss'

import { useState } from 'react'
import { TeamDetailData, addUserToGroups, Group } from '../../services/teamDetail'
import { User } from '../../services/teamMembers'
import { formatDisplayName, getErrorList, getGroupType, removeDuplicateDropdownItems } from '../../utils/utils'
import { DropdownItems } from '../../@types/pageTypes'
import SidebarModal, { SidebarHeader } from '../../components/SidebarModal/SidebarModal'

import { Dialog, Dropdown, Tag } from '@statisticsnorway/ssb-component-library'
import { Skeleton, CircularProgress } from '@mui/material'
import { XCircle } from 'react-feather'

interface AddMember {
  loadingUsers: boolean
  userData: User[] | undefined
  teamDetailData: TeamDetailData | undefined
  teamModalHeader: SidebarHeader
  teamGroups: Group[]
  open: boolean
  onClose: CallableFunction
}

const defaultSelectedUserDropdown = {
  key: 'add-selected-user',
  error: false,
  errorMessage: `Ugyldig navn`,
}
const defaultSelectedUser = {
  id: 'search',
  title: 'Søk ...',
}

const defaultAddUserKey = 'add-user-selected-group'
const defaultSelectedGroup = {
  id: 'velg',
  title: 'Velg ...',
}

const AddTeamMember = ({
  loadingUsers,
  userData,
  teamDetailData,
  teamModalHeader,
  teamGroups,
  open,
  onClose,
}: AddMember) => {
  const [selectedUserDropdown, setSelectedUserDropdown] = useState(defaultSelectedUserDropdown)
  const [selectedUser, setSelectedUser] = useState(defaultSelectedUser)
  const [selectedGroupAddUser, setSelectedGroupAddUser] = useState({
    ...defaultSelectedGroup,
    key: defaultAddUserKey,
  })
  const [teamGroupTags, setTeamGroupTags] = useState<DropdownItems[]>([])
  const [teamGroupTagsError, setTeamGroupTagsError] = useState({
    error: false,
    errorMessage: 'Velg minst én tilgangsgruppe',
  })
  const [addUserToTeamErrors, setAddUserToTeamErrors] = useState<Array<string>>([])
  const [showAddUserSpinner, setShowAddUserSpinner] = useState<boolean>(false)

  const handleAddUser = (item: DropdownItems) => {
    setSelectedUserDropdown({ ...selectedUserDropdown, key: `${defaultSelectedUserDropdown.key}-${item.id}` })
    setSelectedUser(item)
  }

  const handleAddGroupTag = (item: DropdownItems) => {
    const teamGroupsTags = removeDuplicateDropdownItems([...teamGroupTags, item])
    setTeamGroupTags(teamGroupsTags)
    setTeamGroupTagsError({ ...teamGroupTagsError, error: false })
    setSelectedGroupAddUser({ ...item, key: `${defaultAddUserKey}-${item.id}` })
  }

  const handleDeleteGroupTag = (item: DropdownItems) => {
    const teamGroupsTags = teamGroupTags.filter((items) => items !== item)
    setTeamGroupTags(teamGroupsTags)
  }

  const handleAddUserOnSubmit = () => {
    const isSelectedUserValid = selectedUser.id !== 'search'
    if (!isSelectedUserValid) setSelectedUserDropdown({ ...selectedUserDropdown, error: true })
    if (!teamGroupTags.length)
      setTeamGroupTagsError({
        ...teamGroupTagsError,
        error: true,
      })

    if (isSelectedUserValid && teamGroupTags.length) {
      setAddUserToTeamErrors([])
      setShowAddUserSpinner(true)
      addUserToGroups(
        teamGroupTags.map((group) => group.id),
        selectedUser.id
      )
        .then((response) => {
          const errorsList = getErrorList(response)
          if (errorsList.length) {
            setAddUserToTeamErrors(errorsList)
          } else {
            onClose()
            setTeamGroupTags([])
            // Reset fields with their respective keys; re-initializes component
            setSelectedUserDropdown({ ...defaultSelectedUserDropdown })
            setSelectedGroupAddUser({ ...defaultSelectedGroup, key: defaultAddUserKey })
          }
        })
        .catch((e) => setAddUserToTeamErrors(e.message))
        .finally(() => setShowAddUserSpinner(false))
    }
  }

  const renderSidebarModalInfo = (children: JSX.Element) => {
    return (
      <div className={styles.modalBodyDialog}>
        <Dialog type='info'>Det kan ta opp til 45 minutter før personen kan bruke tilgangen</Dialog>
        {children}
      </div>
    )
  }

  const renderSidebarModalWarning = (errorList: string[]) => {
    if (errorList.length) {
      return (
        <Dialog type='warning'>
          {typeof errorList === 'string' ? (
            errorList
          ) : (
            <ul>
              {errorList.map((errors) => (
                <li>{errors}</li>
              ))}
            </ul>
          )}
        </Dialog>
      )
    }
  }

  if (teamDetailData) {
    return (
      <SidebarModal
        open={open}
        onClose={() => onClose()}
        header={teamModalHeader}
        footer={{
          submitButtonText: 'Legg til medlem',
          handleSubmit: handleAddUserOnSubmit,
        }}
        body={{
          modalBodyTitle: 'Legg person til teamet',
          modalBody: (
            <>
              {!loadingUsers ? (
                <Dropdown
                  key={selectedUserDropdown.key}
                  className={styles.inputSpacing}
                  header='Navn'
                  selectedItem={selectedUser}
                  items={userData?.map(({ principal_name, display_name }) => {
                    return {
                      id: principal_name,
                      title: `${formatDisplayName(display_name)} (${principal_name})`,
                    }
                  })}
                  onSelect={(item: DropdownItems) => handleAddUser(item)}
                  error={selectedUserDropdown.error}
                  errorMessage={selectedUserDropdown.errorMessage}
                  searchable
                />
              ) : (
                <div className={styles.inputSpacing}>
                  <Skeleton variant='rectangular' animation='wave' height={65} />
                </div>
              )}
              <Dropdown
                key={selectedGroupAddUser.key}
                className={styles.dropdownSpacing}
                header='Tilgangsgrupper(r)'
                selectedItem={selectedGroupAddUser}
                items={teamGroups.map(({ uniform_name }) => ({
                  id: uniform_name,
                  title: getGroupType(uniform_name),
                }))}
                onSelect={(item: DropdownItems) => handleAddGroupTag(item)}
                error={teamGroupTagsError.error}
                errorMessage={teamGroupTagsError.errorMessage}
              />
              <div className={styles.tagsContainer}>
                {teamGroupTags &&
                  teamGroupTags.map((group) => (
                    <Tag
                      key={`team-group-tag-${group.id}`}
                      icon={<XCircle size={14} />}
                      onClick={() => handleDeleteGroupTag(group)}
                    >
                      {group.title}
                    </Tag>
                  ))}
              </div>
              <div className={styles.modalBodyDialog}>
                {renderSidebarModalInfo(
                  <>
                    {addUserToTeamErrors.length ? renderSidebarModalWarning(addUserToTeamErrors) : null}
                    {showAddUserSpinner && <CircularProgress />}
                  </>
                )}
              </div>
            </>
          ),
        }}
      />
    )
  }
  return
}

export default AddTeamMember
