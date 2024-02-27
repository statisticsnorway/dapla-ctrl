import { useCallback, useEffect, useState } from 'react'
import { Dialog, Text, Link, Tabs, Divider } from '@statisticsnorway/ssb-component-library'

import { TabProps } from '../../@types/pageTypes'
import PageLayout from '../../components/PageLayout/PageLayout'
import Table, { TableData } from '../../components/Table/Table'
import PageSkeleton from '../../components/PageSkeleton/PageSkeleton'

import { fetchAllTeamMembersData, TeamMembersData, User } from '../../services/teamMembers'
import { formatDisplayName } from '../../utils/utils'
import { ApiError } from '../../utils/services'

const TeamMembers = () => {
  const accessToken = localStorage.getItem('access_token') || ''
  const jwt = JSON.parse(atob(accessToken.split('.')[1]))

  const defaultActiveTab = {
    title: 'Mine teammedlemmer',
    path: 'myUsers',
  }

  const [activeTab, setActiveTab] = useState<TabProps | string>(defaultActiveTab)
  const [teamMembersData, setTeamMembersData] = useState<TeamMembersData>()
  const [teamMembersTableData, setTeamMembersTableData] = useState<TableData['data']>()
  const [teamMembersTableTitle, setTeamMembersTableTitle] = useState<string>(defaultActiveTab.title)
  const [error, setError] = useState<ApiError | undefined>()
  const [loading, setLoading] = useState<boolean>(true)

  const prepUserData = useCallback(
    (response: TeamMembersData): TableData['data'] => {
      const teamMember = (activeTab as TabProps)?.path ?? activeTab

      return response[teamMember].users.map((teamMember) => ({
        id: teamMember.principal_name,
        navn: renderUserNameColumn(teamMember),
        team: teamMember.teams.length,
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
  }, [prepUserData, jwt])

  useEffect(() => {
    if (teamMembersData) {
      setTeamMembersTableData(prepUserData(teamMembersData))
    }
  }, [teamMembersData, prepUserData])

  const handleTabClick = (tab: string) => {
    setActiveTab(tab)
    if (tab === 'myUsers') {
      setTeamMembersTableTitle('Mine teammedlemmer')
    } else {
      setTeamMembersTableTitle('Alle teammedlemmer')
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
            activeOnInit={defaultActiveTab.path}
            items={[
              { title: `Mine teammedlemmer (${teamMembersData?.myUsers.users.length ?? 0})`, path: 'myUsers' },
              { title: `Alle teammedlemmer (${teamMembersData?.allUsers.users.length ?? 0})`, path: 'allUsers' },
            ]}
          />
          <Divider dark />
          {teamMembersTableData.length > 0 ? (
            <Table
              title={teamMembersTableTitle}
              columns={teamMembersTableHeaderColumns}
              data={teamMembersTableData as TableData['data']}
            />
          ) : (
            <Dialog type='warning' title='No team members found'>
              You are not a manager in any dapla-team
            </Dialog>
          )}
        </>
      )
    }
  }

  return <PageLayout title={'Teammedlemmer'} content={renderContent()} />
}

export default TeamMembers
