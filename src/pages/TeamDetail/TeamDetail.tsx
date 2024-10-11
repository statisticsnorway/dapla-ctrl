/* eslint-disable react-hooks/exhaustive-deps */
import pageStyles from '../../components/PageLayout/pagelayout.module.scss'
import { TabProps } from '../../@types/pageTypes'

import { useCallback, useContext, useEffect, useState } from 'react'
import PageLayout from '../../components/PageLayout/PageLayout'
import { TeamDetailData, getTeamDetail, Team, SharedBuckets, Group } from '../../services/teamDetail'
import { useParams, useNavigate } from 'react-router-dom'
import { ApiError, TokenData, fetchUserInformationFromAuthToken, isDaplaAdmin } from '../../utils/services'

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
  Link,
  Glossary,
} from '@statisticsnorway/ssb-component-library'
import PageSkeleton from '../../components/PageSkeleton/PageSkeleton'
import { Skeleton } from '@mui/material'

import FormattedTableColumn from '../../components/FormattedTableColumn/FormattedTableColumn'
import { fetchUserSearchData, User } from '../../services/teamMembers'
import AddTeamMember from './AddTeamMember'
import EditTeamMember from './EditTeamMember'
import { AUTONOMY_LEVEL } from '../../content/glossary'

import { Effect } from 'effect'

export interface UserInfo {
  name?: string
  email?: string
  groups?: Group[]
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
      label: 'E-post',
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
    { id: 'antall_personer', label: 'Antall personer', align: 'right' },
  ],
}

const TeamDetail = () => {
  const [activeTab, setActiveTab] = useState<TabProps | string>(TEAM_USERS_TAB)
  const [tokenData, setTokenData] = useState<TokenData>()

  const { setBreadcrumbTeamDetailDisplayName } = useContext(DaplaCtrlContext)
  const [error, setError] = useState<ApiError | undefined>()
  const [refreshData, setRefreshData] = useState<boolean>(false)
  const [loadingTeamData, setLoadingTeamData] = useState<boolean>(true)
  const [loadingUsers, setLoadingUsers] = useState<boolean>(false)
  const [teamDetailData, setTeamDetailData] = useState<TeamDetailData>()
  const [userData, setUserData] = useState<User[]>()
  const [isManager, setIsManager] = useState<boolean>(false)
  const [teamDetailTableTitle, setTeamDetailTableTitle] = useState<string>(TEAM_USERS_TAB.title)
  const [teamDetailTableHeaderColumns, setTeamDetailTableHeaderColumns] = useState<TableData['columns']>(
    TEAM_USERS_TAB.columns
  )
  const [teamDetailTableData, setTeamDetailTableData] = useState<TableData['data']>()

  // Add users to team
  const [openAddUserSidebarModal, setOpenAddUserSidebarModal] = useState<boolean>(false)

  // Edit users in team
  const [openEditUserSidebarModal, setOpenEditUserSidebarModal] = useState<boolean>(false)
  const [editUserInfo, setEditUserInfo] = useState<UserInfo>({ name: '', email: '', groups: [] })

  const { teamId } = useParams<{ teamId: string }>()
  const navigate = useNavigate()
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
                return group.uniform_name.startsWith(baseUniformName)
              })
              .map((group) => getGroupType((response.team as Team).uniform_name, group.uniform_name))
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

  useEffect(() => {
    const checkIsTeamManager = Effect.gen(function* () {
      const teamManagers = (teamDetailData && (teamDetailData.team as Team).managers) ?? []
      let isManager = false
      if (tokenData) {
        const isAdmin = yield* Effect.promise(() => isDaplaAdmin(tokenData.email.toLowerCase()))
        const isManagerResult = teamManagers.some(
          (manager) => manager.principal_name.toLowerCase() === tokenData.email.toLowerCase()
        )
        isManager = isAdmin || isManagerResult
      }
      yield* Effect.sync(() => setIsManager(isManager))
    })

    checkIsTeamManager.pipe(Effect.runPromise)
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
      .finally(() => {
        setLoadingTeamData(false)
        setRefreshData(false)
      })
      .catch((error) => {
        if (error?.code === 404) return navigate('/not-found')
        setError(error as ApiError)
      })
  }, [])

  useEffect(() => {
    if (!teamId || refreshData === false) return
    setLoadingTeamData(true)
    getTeamDetail(teamId)
      .then((response) => {
        const formattedResponse = response as TeamDetailData
        setTeamDetailData(formattedResponse)
        setTeamDetailTableData(prepTeamData(formattedResponse))

        const displayName = formatDisplayName((formattedResponse.team as Team).display_name)
        setBreadcrumbTeamDetailDisplayName({ displayName })
      })
      .finally(() => {
        setLoadingTeamData(false)
        setRefreshData(false)
      })
      .catch((error) => {
        setError(error as ApiError)
      })
  }, [refreshData])

  useEffect(() => {
    if (isManager) {
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
  }, [isManager])

  useEffect(() => {
    if (teamDetailData) {
      if (teamDetailTab === SHARED_BUCKETS_TAB.path) {
        setTeamDetailTableTitle(SHARED_BUCKETS_TAB.title)
        setTeamDetailTableHeaderColumns(SHARED_BUCKETS_TAB.columns)
      } else {
        setTeamDetailTableTitle(TEAM_USERS_TAB.title)
        if (isManager) {
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
      const autonomy_level: string = (teamDetailData.team as Team).autonomy_level || 'UNDEFINED'
      const autonomy_description: string =
        AUTONOMY_LEVEL[autonomy_level]?.text || 'Autonomy level description is undefined'
      return (
        <>
          <LeadParagraph className={pageStyles.description}>
            <Text medium className={pageStyles.descriptionSpacing}>
              {(teamDetailData.team as Team).uniform_name ?? ''}
            </Text>
            <Text medium>{formatDisplayName((teamDetailData.team as Team).section_manager.display_name ?? '')}</Text>
            <Text medium>{(teamDetailData.team as Team).section_name ?? ''}</Text>
          </LeadParagraph>
          <Text>
            <Glossary explanation={autonomy_description}>{autonomy_level}</Glossary>
          </Text>
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
  return (
    <>
      <AddTeamMember
        loadingUsers={loadingUsers}
        setRefreshData={setRefreshData}
        userData={userData}
        teamDetailData={teamDetailData}
        teamModalHeader={teamModalHeader}
        teamGroups={teamGroups}
        open={openAddUserSidebarModal}
        onClose={() => setOpenAddUserSidebarModal(false)}
      />
      <EditTeamMember
        editUserInfo={editUserInfo}
        setRefreshData={setRefreshData}
        teamDetailData={teamDetailData}
        teamModalHeader={teamModalHeader}
        teamGroups={teamGroups}
        open={openEditUserSidebarModal}
        onClose={() => setOpenEditUserSidebarModal(false)}
      />
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
          isManager && teamDetailData ? (
            <Button onClick={() => setOpenAddUserSidebarModal(true)}>+ Nytt medlem</Button>
          ) : undefined
        }
      />
    </>
  )
}

export default TeamDetail
