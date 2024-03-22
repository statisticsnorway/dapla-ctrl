/* eslint-disable react-hooks/exhaustive-deps */
import pageStyles from '../../components/PageLayout/pagelayout.module.scss'
import styles from './teamDetail.module.scss'

import { DropdownItems, TabProps } from '../../@types/pageTypes'

import { ReactElement, useCallback, useContext, useEffect, useState } from 'react'
import PageLayout from '../../components/PageLayout/PageLayout'
import {
  TeamDetailData,
  getTeamDetail,
  Team,
  SharedBuckets,
  addUserToGroups,
  removeUserFromGroups,
  Group,
  JobResponse,
} from '../../services/teamDetail'
import { useParams } from 'react-router-dom'
import { ApiError, TokenData, fetchUserInformationFromAuthToken } from '../../utils/services'

import { DaplaCtrlContext } from '../../provider/DaplaCtrlProvider'
import Table, { TableData } from '../../components/Table/Table'
import { formatDisplayName, getGroupType } from '../../utils/utils'
import {
  Text,
  Dialog,
  LeadParagraph,
  Divider,
  Tabs,
  Button,
  Dropdown,
  Tag,
  Link,
} from '@statisticsnorway/ssb-component-library'
import PageSkeleton from '../../components/PageSkeleton/PageSkeleton'
import { Skeleton, CircularProgress } from '@mui/material'
import { XCircle } from 'react-feather'
import FormattedTableColumn from '../../components/FormattedTableColumn'
import SidebarModal from '../../components/SidebarModal/SidebarModal'
import DeleteLink from '../../components/DeleteLink/DeleteLink'
import { fetchUserSearchData, User } from '../../services/teamMembers'

interface UserInfo {
  name?: string
  email?: string
  groups?: Group[]
}

interface EditUserStates {
  [key: string]: boolean | Array<string>
}

const TEAM_USERS_TAB = {
  title: 'Teammedlemmer',
  path: 'team',
  columns: [
    {
      id: 'navn',
      label: 'Navn',
    },
    {
      id: 'gruppe',
      label: 'Gruppe',
    },
    {
      id: 'epost',
      label: 'Epost',
    },
  ],
}

const SHARED_BUCKETS_TAB = {
  title: 'Delte data',
  path: 'sharedBuckets',
  columns: [
    {
      id: 'navn',
      label: 'Navn',
    },
    {
      id: 'tilgang',
      label: 'Tilgang',
    },
    // { id: 'delte_data', label: 'Delte data' },
    { id: 'antall_personer', label: 'Antall personer' },
  ],
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
const defaultEditUserKey = 'edit-user-selected-group'
const defaultSelectedGroup = {
  id: 'velg',
  title: 'Velg ...',
}

/* TODO: States that we want to keep the history of when editing user
 * Selected group */

const TeamDetail = () => {
  const [activeTab, setActiveTab] = useState<TabProps | string>(TEAM_USERS_TAB)
  const [tokenData, setTokenData] = useState<TokenData>()

  const { setBreadcrumbTeamDetailDisplayName } = useContext(DaplaCtrlContext)
  const [error, setError] = useState<ApiError | undefined>()
  const [loadingTeamData, setLoadingTeamData] = useState<boolean>(true)
  const [loadingUsers, setLoadingUsers] = useState<boolean>(false)
  const [teamDetailData, setTeamDetailData] = useState<TeamDetailData>()
  const [userData, setUserData] = useState<User[]>()
  const [teamDetailTableTitle, setTeamDetailTableTitle] = useState<string>(TEAM_USERS_TAB.title)
  const [teamDetailTableHeaderColumns, setTeamDetailTableHeaderColumns] = useState<TableData['columns']>(
    TEAM_USERS_TAB.columns
  )
  const [teamDetailTableData, setTeamDetailTableData] = useState<TableData['data']>()

  // Add users to team
  const [openAddUserSidebarModal, setOpenAddUserSidebarModal] = useState<boolean>(false)
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

  // Edit users in team
  const [openEditUserSidebarModal, setOpenEditUserSidebarModal] = useState<boolean>(false)
  const [editUserInfo, setEditUserInfo] = useState<UserInfo>({ name: '', email: '', groups: [] })
  const [selectedGroupEditUser, setSelectedGroupEditUser] = useState({
    ...defaultSelectedGroup,
    key: defaultEditUserKey,
  })
  const [userGroupTags, setUserGroupTags] = useState<DropdownItems[]>([])
  const [editUserErrors, setEditUserErrors] = useState<EditUserStates>({})
  const [showEditUserSpinner, setShowEditUserSpinner] = useState<EditUserStates>({})

  const { teamId } = useParams<{ teamId: string }>()
  const teamDetailTab = (activeTab as TabProps)?.path ?? activeTab

  const prepTeamData = useCallback(
    (response: TeamDetailData): TableData['data'] => {
      const sharedBucketsTab = SHARED_BUCKETS_TAB.path
      if (teamDetailTab === sharedBucketsTab) {
        const sharedBuckets = (response[sharedBucketsTab] as SharedBuckets).items
        if (!sharedBuckets) return []

        return sharedBuckets.map(({ short_name, bucket_name, metrics }) => {
          const teams_count = metrics?.teams_count
          return {
            id: short_name,
            navn: <FormattedTableColumn href={`/${teamId}/${short_name}`} linkText={short_name} text={bucket_name} />,
            tilgang: typeof teams_count === 'number' ? `${teams_count} team` : teams_count,
            // delte_data: '-', // To be implemented; data does not exist in the API yet.
            antall_personer: metrics?.users_count,
          }
        })
      } else {
        const teamUsers = (response[TEAM_USERS_TAB.path] as Team).users
        if (!teamUsers) return []

        return teamUsers.map(({ display_name, principal_name, section_name, groups }) => {
          const userFullName = formatDisplayName(display_name)
          const userGroups = groups?.filter((group) =>
            group.uniform_name.startsWith((response.team as Team).uniform_name)
          ) as Group[]
          return {
            id: userFullName,
            navn: (
              <FormattedTableColumn
                href={`/teammedlemmer/${principal_name}`}
                linkText={formatDisplayName(display_name)}
                text={section_name}
              />
            ),
            seksjon: section_name, // Makes section name searchable and sortable in table by including the field
            gruppe: groups
              ?.filter((group) => {
                const baseUniformName = (response.team as Team).uniform_name
                const allowedSuffixes = ['-managers', '-developers', '-data-admins', '-support', '-consumers']
                if (group.uniform_name.startsWith(baseUniformName)) {
                  const suffix = group.uniform_name.slice(baseUniformName.length)
                  return allowedSuffixes.some((allowedSuffix) => suffix.startsWith(allowedSuffix))
                }
                return false
              })
              .map((group) => getGroupType(group.uniform_name))
              .join(', '),
            epost: principal_name,
            editUser: (
              <span>
                <Link
                  onClick={() => {
                    setOpenEditUserSidebarModal(true)
                    setEditUserInfo({
                      name: formatDisplayName(display_name),
                      email: principal_name,
                      groups: userGroups,
                    })
                    setUserGroupTags(
                      userGroups.map(({ uniform_name }) => {
                        return { id: uniform_name, title: getGroupType(uniform_name) }
                      })
                    )
                  }}
                >
                  Endre
                </Link>
              </span>
            ),
          }
        })
      }
    },
    [activeTab]
  )

  const isTeamManager = useCallback(() => {
    const teamManagers = (teamDetailData && (teamDetailData.team as Team).managers) ?? []
    return teamManagers?.some((manager) => manager.principal_name === tokenData?.email)
  }, [tokenData, teamDetailData])

  useEffect(() => {
    if (!teamId) return
    fetchUserInformationFromAuthToken()
      .then((tokenData) => setTokenData(tokenData))
      .catch((error) => setError(error as ApiError))
    getTeamDetail(teamId)
      .then((response) => {
        const formattedResponse = response as TeamDetailData
        setTeamDetailData(formattedResponse)
        setTeamDetailTableData(prepTeamData(formattedResponse))

        const displayName = formatDisplayName((formattedResponse.team as Team).display_name)
        setBreadcrumbTeamDetailDisplayName({ displayName })
      })
      .finally(() => setLoadingTeamData(false))
      .catch((error) => {
        setError(error as ApiError)
      })
  }, [])

  useEffect(() => {
    if (isTeamManager()) {
      setTeamDetailTableHeaderColumns([
        ...TEAM_USERS_TAB.columns,
        {
          id: 'editUser',
          label: '',
          unsortable: true,
          align: 'center',
        },
      ])
    }
  }, [isTeamManager])

  useEffect(() => {
    if (teamDetailData) {
      if (teamDetailTab === SHARED_BUCKETS_TAB.path) {
        setTeamDetailTableTitle(SHARED_BUCKETS_TAB.title)
        setTeamDetailTableHeaderColumns(SHARED_BUCKETS_TAB.columns)
      } else {
        setTeamDetailTableTitle(TEAM_USERS_TAB.title)
        if (isTeamManager()) {
          setTeamDetailTableHeaderColumns([
            ...TEAM_USERS_TAB.columns,
            {
              id: 'editUser',
              label: '',
              unsortable: true,
              align: 'center',
            },
          ])
        } else {
          setTeamDetailTableHeaderColumns(TEAM_USERS_TAB.columns)
        }
      }
      setTeamDetailTableData(prepTeamData(teamDetailData))
    }
  }, [prepTeamData])

  const getUsersAutoCompleteData = () => {
    if (userData) return
    setLoadingUsers(true)
    fetchUserSearchData()
      .then((users) => {
        const filteredUsers = users.filter(
          (allUsers) =>
            !(teamDetailData?.team as Team).users?.some(
              (teamUsers) => allUsers.principal_name === teamUsers.principal_name
            )
        )
        setUserData(filteredUsers)
      })
      .catch((error) => {
        setError(error as ApiError)
      })
      .finally(() => setLoadingUsers(false))
  }

  useEffect(() => {
    if (openAddUserSidebarModal) {
      getUsersAutoCompleteData()
    }
  }, [openAddUserSidebarModal])

  const handleTabClick = (tab: string) => setActiveTab(tab)

  const renderErrorAlert = () => {
    return (
      <Dialog type='warning' title='Could not fetch teams'>
        {`${error?.code} - ${error?.message}`}
      </Dialog>
    )
  }

  const renderContent = () => {
    if (error) return renderErrorAlert()
    if (loadingTeamData) return <PageSkeleton hasDescription />

    if (teamDetailData && teamDetailTableHeaderColumns && teamDetailTableData) {
      return (
        <>
          <LeadParagraph className={pageStyles.description}>
            <Text medium className={pageStyles.descriptionSpacing}>
              {(teamDetailData.team as Team).uniform_name ?? ''}
            </Text>
            <Text medium>{formatDisplayName((teamDetailData.team as Team).section_manager.display_name ?? '')}</Text>
            <Text medium>{(teamDetailData.team as Team).section_name ?? ''}</Text>
          </LeadParagraph>
          <Tabs
            onClick={handleTabClick}
            activeOnInit={TEAM_USERS_TAB.path}
            items={[
              {
                title: `${TEAM_USERS_TAB.title} (${(teamDetailData?.team as Team).users?.length ?? 0})`,
                path: TEAM_USERS_TAB.path,
              },
              {
                title: `${SHARED_BUCKETS_TAB.title} (${(teamDetailData?.sharedBuckets as SharedBuckets).items?.length ?? 0})`,
                path: SHARED_BUCKETS_TAB.path,
              },
            ]}
          />
          <Divider dark />
          <Table
            title={teamDetailTableTitle}
            columns={teamDetailTableHeaderColumns}
            data={teamDetailTableData as TableData['data']}
          />
        </>
      )
    }
  }

  const handleAddUser = (item: DropdownItems) => {
    setSelectedUserDropdown({ ...selectedUserDropdown, key: `${defaultSelectedUserDropdown.key}-${item.id}` })
    setSelectedUser(item)
  }

  const removeDuplicateDropdownItems = (items: DropdownItems[]) => {
    return items.reduce((acc: DropdownItems[], dropdownItem: DropdownItems) => {
      const ids = acc.map((obj) => obj.id)
      if (!ids.includes(dropdownItem.id)) {
        acc.push(dropdownItem)
      }
      return acc
    }, [])
  }

  const handleAddGroupTag = (item: DropdownItems) => {
    if (openAddUserSidebarModal) {
      const teamGroupsTags = removeDuplicateDropdownItems([...teamGroupTags, item])
      setTeamGroupTags(teamGroupsTags)
      setTeamGroupTagsError({ ...teamGroupTagsError, error: false })
      setSelectedGroupAddUser({ ...item, key: `${defaultAddUserKey}-${item.id}` })
    }

    if (openEditUserSidebarModal) {
      const userGroupsTagsList = removeDuplicateDropdownItems([...userGroupTags, item])
      setUserGroupTags(userGroupsTagsList)
      setSelectedGroupEditUser({ ...item, key: `${defaultEditUserKey}-${item.id}` })
    }
  }

  const handleDeleteGroupTag = (item: DropdownItems) => {
    if (openAddUserSidebarModal) {
      const teamGroupsTags = teamGroupTags.filter((items) => items !== item)
      setTeamGroupTags(teamGroupsTags)
    }

    if (openEditUserSidebarModal) {
      const userGroupsTags = userGroupTags.filter((items) => items !== item)
      setUserGroupTags(userGroupsTags)
    }
  }

  const getErrorList = (response: JobResponse[]) => {
    return response
      .map(({ status, detail }) => {
        if ((detail && status === 'ERROR') || (detail && status === 'IGNORED')) {
          return detail
        }
        return ''
      })
      .filter((str) => str !== '')
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
            setOpenAddUserSidebarModal(false)
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

  const resetEditUserValues = () => {
    setEditUserErrors({ ...editUserErrors, [`${editUserInfo.email}`]: [] })
    setShowEditUserSpinner({ ...showEditUserSpinner, [`${editUserInfo.email}`]: true })
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
            setOpenEditUserSidebarModal(false)
            // Reset fields with their respective keys; re-initializes component
            setSelectedGroupEditUser({ ...defaultSelectedGroup, key: defaultEditUserKey })
          }
        })
        .catch((e) => setEditUserErrors({ ...editUserErrors, [`${editUserInfo.email}`]: e.message }))
        .finally(() => setShowEditUserSpinner({ ...showEditUserSpinner, [`${editUserInfo.email}`]: false }))

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
            setOpenEditUserSidebarModal(false)
            // Reset fields with their respective keys; re-initializes component
            setSelectedGroupEditUser({ ...defaultSelectedGroup, key: defaultEditUserKey })
          }
        })
        .catch((e) => setEditUserErrors({ ...editUserErrors, [`${editUserInfo.email}`]: e.message }))
        .finally(() => setShowEditUserSpinner({ ...showEditUserSpinner, [`${editUserInfo.email}`]: false }))

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
            setOpenEditUserSidebarModal(false)
            // Reset fields with their respective keys; re-initializes component
            setSelectedGroupEditUser({ ...defaultSelectedGroup, key: defaultEditUserKey })
          }
        })
        .catch((e) => setEditUserErrors({ ...editUserErrors, [`${editUserInfo.email}`]: e.message }))
        .finally(() => setShowEditUserSpinner({ ...showEditUserSpinner, [`${editUserInfo.email}`]: false }))

      return
    }
  }

  const handleDeleteUser = () => {
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
            setOpenEditUserSidebarModal(false)
            // Reset fields with their respective keys; re-initializes component
            setSelectedGroupEditUser({ ...defaultSelectedGroup, key: defaultEditUserKey })
          }
        })
        .catch((e) => setEditUserErrors({ ...editUserErrors, [`${editUserInfo.email}`]: e.message }))
        .finally(() => setShowEditUserSpinner({ ...showEditUserSpinner, [`${editUserInfo.email}`]: false }))

      return
    }
  }

  const renderSidebarModalInfo = (children: ReactElement) => {
    return (
      <div className={styles.modalBodyDialog}>
        <Dialog type='info'>Det kan ta opp til 45 minutter før personen kan bruke tilgangen</Dialog>
        {children}
      </div>
    )
  }

  const renderSidebarModalWarning = (errorList: string[]) => {
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

  const teamModalHeader = teamDetailData
    ? {
        modalType: 'Medlem',
        modalTitle: `${(teamDetailData?.team as Team).display_name}`,
        modalDescription: `${(teamDetailData?.team as Team).uniform_name}`,
      }
    : {
        modalTitle: '',
      }
  const teamGroups = teamDetailData ? ((teamDetailData.team as Team).groups as Group[]) : []
  const renderAddUserSidebarModal = () => {
    if (teamDetailData) {
      return (
        <SidebarModal
          open={openAddUserSidebarModal}
          onClose={() => setOpenAddUserSidebarModal(false)}
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
                        title: formatDisplayName(display_name),
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
  }

  const renderEditUserSidebarModal = () => {
    if (teamDetailData && editUserInfo) {
      return (
        <SidebarModal
          open={openEditUserSidebarModal}
          onClose={() => setOpenEditUserSidebarModal(false)}
          header={teamModalHeader}
          footer={{
            submitButtonText: 'Oppdater Tilgang',
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
                  items={teamGroups.map(({ uniform_name }) => ({
                    id: uniform_name,
                    title: getGroupType(uniform_name),
                  }))}
                  onSelect={(item: DropdownItems) => handleAddGroupTag(item)}
                />
                <div className={styles.tagsContainer}>
                  {userGroupTags &&
                    userGroupTags.map((group) => (
                      <Tag
                        key={`user-group-tag-${group.id}`}
                        icon={<XCircle size={14} />}
                        onClick={() => handleDeleteGroupTag(group)}
                      >
                        {group.title}
                      </Tag>
                    ))}
                </div>
                <div className={styles.modalBodyDialog}>
                  <DeleteLink handleDeleteUser={handleDeleteUser} icon>
                    Fjern fra teamet
                  </DeleteLink>
                  {renderSidebarModalInfo(
                    <>
                      {editUserErrors?.[editUserInfo.email as string]
                        ? renderSidebarModalWarning(editUserErrors?.[editUserInfo.email as string] as string[])
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

  return (
    <>
      {renderAddUserSidebarModal()}
      {renderEditUserSidebarModal()}
      <PageLayout
        title={
          !loadingTeamData && teamDetailData ? (
            (teamDetailData.team as Team).display_name
          ) : (
            <Skeleton variant='rectangular' animation='wave' width={350} height={90} />
          )
        }
        content={renderContent()}
        button={
          isTeamManager() ? <Button onClick={() => setOpenAddUserSidebarModal(true)}>+ Nytt medlem</Button> : undefined
        }
      />
    </>
  )
}

export default TeamDetail
