import styles from './teamDetail.module.scss'

import { useState, useEffect } from 'react'
import { UserInfo } from './TeamDetail'
import {
  TeamDetailData,
  Team,
  Group,
  addUserToGroups,
  removeUserFromGroups,
  JobResponse,
} from '../../services/teamDetail'
import { DropdownItem } from '../../@types/pageTypes'
import { getErrorList, getGroupType, removeDuplicateDropdownItems } from '../../utils/utils'
import SidebarModal, { SidebarHeader } from '../../components/SidebarModal/SidebarModal'
import Modal from '../../components/Modal/Modal'
import DeleteLink from '../../components/DeleteLink/DeleteLink'
import { renderSidebarModalInfo, renderSidebarModalWarning } from './teamDetailDialog'
import { displayGroupItem, standardGroups } from './common'

import { Dialog, Dropdown, Tag, Link, Button } from '@statisticsnorway/ssb-component-library'
import { CircularProgress } from '@mui/material'
import { XCircle, Trash2 } from 'react-feather'

interface EditUserStates {
  [key: string]: boolean | string[] | JobResponse[]
}

interface EditTeamMember {
  editUserInfo: UserInfo
  setRefreshData: React.Dispatch<React.SetStateAction<boolean>>
  teamDetailData: TeamDetailData | undefined
  teamModalHeader: SidebarHeader
  teamGroups: Group[]
  open: boolean
  onClose: CallableFunction
}

const defaultEditUserKey = 'edit-user-selected-group'
const defaultSelectedGroup = {
  id: 'velg',
  title: 'Velg ...',
}

const EditTeamMember = ({
  editUserInfo,
  setRefreshData,
  teamDetailData,
  teamModalHeader,
  teamGroups,
  open,
  onClose,
}: EditTeamMember) => {
  const [selectedGroupEditUser, setSelectedGroupEditUser] = useState({
    ...defaultSelectedGroup,
    key: defaultEditUserKey,
  })
  const [userGroupTags, setUserGroupTags] = useState<DropdownItem[]>([])
  const [editUserErrors, setEditUserErrors] = useState<EditUserStates>({})
  const [showEditUserSpinner, setShowEditUserSpinner] = useState<EditUserStates>({})

  const [openDeleteUserConfirmation, setOpenDeleteUserConfirmation] = useState<boolean>(false)
  const [managedGroupsWarning, setManagedGroupsWarning] = useState(false)

  useEffect(() => {
    if (teamDetailData) {
      const userGroups = editUserInfo.groups?.filter((group) =>
        group.uniform_name.startsWith((teamDetailData.team as Team).uniform_name)
      ) as Group[]
      setUserGroupTags(
        userGroups.map(({ uniform_name }) => {
          const groupType = getGroupType((teamDetailData['team'] as Team).uniform_name, uniform_name)
          return { id: uniform_name, title: groupType, disabled: !standardGroups.includes(groupType) }
        })
      )
      setSelectedGroupEditUser({
        ...defaultSelectedGroup,
        key: `${defaultEditUserKey}-${editUserInfo.email}`,
      })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [editUserInfo])

  const handleAddGroupTag = (item: DropdownItem) => {
    const userGroupsTagsList = removeDuplicateDropdownItems([...userGroupTags, item])
    setUserGroupTags(userGroupsTagsList)
    setSelectedGroupEditUser({ ...item, key: `${defaultEditUserKey}-${item.id}` })
  }

  const handleDeleteGroupTag = (item: DropdownItem) => {
    const userGroupsTags = userGroupTags.filter((items) => items !== item)
    setUserGroupTags(userGroupsTags)
  }

  const resetEditUserValues = () => {
    setEditUserErrors({ ...editUserErrors, [`${editUserInfo.email}`]: [] })
    setShowEditUserSpinner({ ...showEditUserSpinner, [`${editUserInfo.email}`]: true })
  }

  const closeModalAndResetSelectedGroup = () => {
    onClose()
    // Reset fields with their respective keys; re-initializes component
    setSelectedGroupEditUser({ ...defaultSelectedGroup, key: defaultEditUserKey })
  }

  const handleEditUserOnSubmit = () => {
    const addedGroups =
      userGroupTags?.filter((groupTag) => !editUserInfo.groups?.some((group) => groupTag.id === group.uniform_name)) ??
      []
    const removedGroups =
      editUserInfo.groups?.filter((group) => !userGroupTags?.some((groupTag) => groupTag.id === group.uniform_name)) ??
      []

    if ((addedGroups.length && removedGroups.length) || addedGroups.length || removedGroups.length) {
      resetEditUserValues()
    }

    if (addedGroups.length && removedGroups.length) {
      Promise.all([
        addUserToGroups(
          addedGroups.map((group) => group.id),
          editUserInfo?.email as string
        ),
        removeUserFromGroups(
          removedGroups.map((group) => group.uniform_name),
          editUserInfo?.email as string
        ),
      ])
        .then((response) => {
          const flattenedResponse = [...response[0], ...response[1]]
          const errorsList = getErrorList(flattenedResponse)
          if (errorsList.length) {
            setEditUserErrors({ ...editUserErrors, [`${editUserInfo.email}`]: errorsList })
          } else {
            closeModalAndResetSelectedGroup()
          }
        })
        .catch((e) => setEditUserErrors({ ...editUserErrors, [`${editUserInfo.email}`]: e.message }))
        .finally(() => {
          setShowEditUserSpinner({ ...showEditUserSpinner, [`${editUserInfo.email}`]: false })
          setRefreshData(true)
        })

      return
    }

    if (addedGroups.length) {
      addUserToGroups(
        addedGroups.map((group) => group.id),
        editUserInfo?.email as string
      )
        .then((response) => {
          const errorsList = getErrorList(response)
          if (errorsList.length) {
            setEditUserErrors({ ...editUserErrors, [`${editUserInfo.email}`]: errorsList })
          } else {
            closeModalAndResetSelectedGroup()
          }
        })
        .catch((e) => setEditUserErrors({ ...editUserErrors, [`${editUserInfo.email}`]: e.message }))
        .finally(() => {
          setShowEditUserSpinner({ ...showEditUserSpinner, [`${editUserInfo.email}`]: false })
          setRefreshData(true)
        })

      return
    }

    if (removedGroups.length) {
      removeUserFromGroups(
        removedGroups?.map((group) => group.uniform_name),
        editUserInfo.email as string
      )
        .then((response) => {
          const errorsList = getErrorList(response)
          if (errorsList.length) {
            setEditUserErrors({ ...editUserErrors, [`${editUserInfo.email}`]: errorsList })
          } else {
            closeModalAndResetSelectedGroup()
          }
        })
        .catch((e) => setEditUserErrors({ ...editUserErrors, [`${editUserInfo.email}`]: e.message }))
        .finally(() => {
          setShowEditUserSpinner({ ...showEditUserSpinner, [`${editUserInfo.email}`]: false })
          setRefreshData(true)
        })
      return
    }
  }

  const handleDeleteUser = () => {
    setOpenDeleteUserConfirmation(false)

    if (editUserInfo.groups && editUserInfo.groups.length) {
      resetEditUserValues()

      removeUserFromGroups(
        editUserInfo.groups.map(({ uniform_name }) => uniform_name),
        editUserInfo.email as string
      )
        .then((response) => {
          const errorsList = getErrorList(response)
          if (errorsList.length) {
            setEditUserErrors({ ...editUserErrors, [`${editUserInfo.email}`]: errorsList })
          } else {
            closeModalAndResetSelectedGroup()
          }
        })
        .catch((e) => setEditUserErrors({ ...editUserErrors, [`${editUserInfo.email}`]: e.message }))
        .finally(() => {
          setShowEditUserSpinner({ ...showEditUserSpinner, [`${editUserInfo.email}`]: false })
          setRefreshData(true)
        })
      return
    }
  }

  const renderDeleteUserConfirmationModal = () => {
    return (
      <Modal
        open={openDeleteUserConfirmation}
        onClose={() => setOpenDeleteUserConfirmation(false)}
        modalTitle={
          <>
            <Trash2 size={24} />
            Fjern tilgang
          </>
        }
        body={
          <>{`Er du sikker på at du vil fjerne "${editUserInfo.name}" fra ${teamDetailData ? (teamDetailData?.team as Team).display_name : ''}?`}</>
        }
        footer={
          <>
            <span>
              <Link onClick={() => setOpenDeleteUserConfirmation(false)}>Avbryt</Link>
            </span>
            <Button onClick={handleDeleteUser} primary>
              Fjern
            </Button>
          </>
        }
      />
    )
  }

  if (teamDetailData && editUserInfo) {
    return (
      <SidebarModal
        open={open}
        onClose={() => {
          setManagedGroupsWarning(false)
          onClose()
        }}
        header={teamModalHeader}
        footer={{
          submitButtonText: 'Oppdater tilgang',
          handleSubmit: handleEditUserOnSubmit,
        }}
        body={{
          modalBodyTitle: `Endre tilgang til "${editUserInfo.name}"`,
          modalBody: (
            <>
              <Dropdown
                key={selectedGroupEditUser.key}
                className={styles.dropdownSpacing}
                header='Tilgangsgrupper(r)'
                selectedItem={selectedGroupEditUser}
                items={teamGroups.map(displayGroupItem(teamDetailData['team'] as Team))}
                onSelect={(item: DropdownItem) => handleAddGroupTag(item)}
              />
              <div className={styles.tagsContainer}>
                {userGroupTags &&
                  userGroupTags.map((group: DropdownItem) => (
                    <Tag
                      key={`user-group-tag-${group.id}`}
                      icon={<XCircle size={14} />}
                      onClick={() => (group.disabled ? setManagedGroupsWarning(true) : handleDeleteGroupTag(group))}
                    >
                      {group.title}
                    </Tag>
                  ))}
              </div>
              {managedGroupsWarning && (
                <Dialog type='warning' title='Bare standard grupper kan fjernes'>
                  Dapla Ctrl støtter ikke å fjerne andre grupper enn "data-admins", "developers" og "managers".
                </Dialog>
              )}
              <div className={styles.modalBodyDialog}>
                <DeleteLink handleDeleteUser={() => setOpenDeleteUserConfirmation(true)} icon>
                  Fjern fra teamet
                </DeleteLink>
                {renderDeleteUserConfirmationModal()}
                {renderSidebarModalInfo(
                  <>
                    {editUserErrors?.[editUserInfo.email as string]
                      ? renderSidebarModalWarning(editUserErrors?.[editUserInfo.email as string] as JobResponse[])
                      : null}
                    {showEditUserSpinner?.[editUserInfo.email as string] && <CircularProgress />}
                  </>
                )}
              </div>
            </>
          ),
        }}
      />
    )
  }
}

export default EditTeamMember
