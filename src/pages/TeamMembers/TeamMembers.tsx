/* eslint-disable react-hooks/exhaustive-deps */
import { useCallback, useEffect, useState } from 'react'
import { Dialog, Text, Link, Tabs, Divider } from '@statisticsnorway/ssb-component-library'

import { TabProps } from '../../@types/pageTypes'
import PageLayout from '../../components/PageLayout/PageLayout'
import Table, { TableData } from '../../components/Table/Table'
import PageSkeleton from '../../components/PageSkeleton/PageSkeleton'

import { fetchAllTeamMembersData, TeamMembersData, User } from '../../services/teamMembers'
import { formatDisplayName } from '../../utils/utils'
import { ApiError } from '../../utils/services'

const MY_USERS_TAB = {
  title: 'Mine teammedlemmer',
  path: 'myUsers',
}

const ALL_USERS_TAB = {
  title: 'Alle teammedlemmer',
  path: 'allUsers',
}

const TeamMembers = () => {
  const accessToken = localStorage.getItem('access_token') || ''
  const jwt = JSON.parse(atob(accessToken.split('.')[1]))

  const [activeTab, setActiveTab] = useState<TabProps | string>(MY_USERS_TAB)
  const [teamMembersData, setTeamMembersData] = useState<TeamMembersData>()
  const [teamMembersTableData, setTeamMembersTableData] = useState<TableData['data']>()
  const [teamMembersTableTitle, setTeamMembersTableTitle] = useState<string>(MY_USERS_TAB.title)
  const [error, setError] = useState<ApiError | undefined>()
  const [loading, setLoading] = useState<boolean>(true)

  const prepUserData = useCallback(
    (response: TeamMembersData): TableData['data'] => {
      const teamMember = (activeTab as TabProps)?.path ?? activeTab

      return response[teamMember].users.map((teamMember) => ({
        id: formatDisplayName(teamMember.display_name),
        navn: renderUserNameColumn(teamMember),
        team: teamMember.teams.length,
        epost: teamMember.principal_name,
        data_admin_roller: teamMember.groups.filter((group) => group.uniform_name.endsWith('data-admins')).length,
        seksjon: teamMember.section_name, // Makes section name searchable and sortable in table by including the field
        seksjonsleder: formatDisplayName(
          teamMember.section_manager && teamMember.section_manager.length > 0
            ? teamMember.section_manager[0].display_name
            : 'Seksjonsleder ikke funnet'
        ),
      }))
    },
    [activeTab]
  )

  useEffect(() => {
    if (!jwt) return
    fetchAllTeamMembersData(jwt.email)
      .then((response) => {
        setTeamMembersData(response as TeamMembersData)
        setTeamMembersTableData(prepUserData(response as TeamMembersData))
      })
      .finally(() => setLoading(false))
      .catch((error) => {
        setError(error as ApiError)
      })
  }, [])

  useEffect(() => {
    if (teamMembersData) {
      setTeamMembersTableData(prepUserData(teamMembersData)) // Update Table view on Tab onClick
    }
  }, [prepUserData])

  const handleTabClick = (tab: string) => {
    setActiveTab(tab)
    if (tab === MY_USERS_TAB.path) {
      setTeamMembersTableTitle(MY_USERS_TAB.title)
    } else {
      setTeamMembersTableTitle(ALL_USERS_TAB.title)
    }
  }

  const renderUserNameColumn = (user: User) => {
    return (
      <>
        <span>
          <Link href={`/teammedlemmer/${user.principal_name}`}>
            <b>{user.display_name}</b>
          </Link>
        </span>
        {user.section_name && <Text>{user.section_name}</Text>}
      </>
    )
  }

  const renderErrorAlert = () => {
    return (
      <Dialog type='warning' title='Could not fetch users'>
        {`${error?.code} - ${error?.message}`}
      </Dialog>
    )
  }

  const renderContent = () => {
    if (error) return renderErrorAlert()
    if (loading) return <PageSkeleton />

    if (teamMembersTableData) {
      const teamMembersTableHeaderColumns = [
        {
          id: 'navn',
          label: 'Navn',
        },
        {
          id: 'team',
          label: 'Team',
        },
        {
          id: 'data_admin_roller',
          label: 'Data-admin-roller',
        },
        {
          id: 'seksjonsleder',
          label: 'Seksjonsleder',
        },
      ]

      return (
        <>
          <Tabs
            onClick={handleTabClick}
            activeOnInit={MY_USERS_TAB.path}
            items={[
              {
                title: `${MY_USERS_TAB.title} (${teamMembersData?.myUsers.users.length ?? 0})`,
                path: MY_USERS_TAB.path,
              },
              {
                title: `${ALL_USERS_TAB.title} (${teamMembersData?.allUsers.users.length ?? 0})`,
                path: ALL_USERS_TAB.path,
              },
            ]}
          />
          <Divider dark />
          <Table
            title={teamMembersTableTitle}
            columns={teamMembersTableHeaderColumns}
            data={teamMembersTableData as TableData['data']}
          />
        </>
      )
    }
  }

  return <PageLayout title={'Teammedlemmer'} content={renderContent()} />
}

export default TeamMembers
