import pageLayoutStyles from '../../components/PageLayout/pagelayout.module.scss'

import { useCallback, useEffect, useState } from 'react'
import { Dialog, Title, Text, Link, Tabs, Divider } from '@statisticsnorway/ssb-component-library'

import { TabProps } from '../../@types/pageTypes'
import PageLayout from '../../components/PageLayout/PageLayout'
import Table, { TableData } from '../../components/Table/Table'
import PageSkeleton from '../../components/PageSkeleton/PageSkeleton'

import { ErrorResponse } from '../../@types/error'

import { fetchAllTeamMembersData, TeamMembersData, User } from '../../services/teamMembers'
import { formatDisplayName } from '../../utils/utils'

export default function TeamMembers() {
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
  const [error, setError] = useState<ErrorResponse | undefined>()
  const [loading, setLoading] = useState<boolean>(true)

  const prepUserData = useCallback(
    (response: TeamMembersData): TableData['data'] => {
      const teamMember = (activeTab as TabProps)?.path ?? activeTab

      return response[teamMember].users.map((teamMember) => ({
        id: teamMember.principal_name,
        navn: renderUserNameColumn(teamMember),
        team: teamMember.teams.length,
        data_admin_roller: teamMember.groups.filter((group) => group.uniform_name.endsWith('data-admins')).length,
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
        if ((response as ErrorResponse).error) {
          setError(response as ErrorResponse)
        } else {
          setTeamMembersData(response as TeamMembersData)
          setTeamMembersTableData(prepUserData(response as TeamMembersData))
        }
      })
      .finally(() => setLoading(false))
      .catch((error) => {
        setError(error.toString())
      })
  }, [prepUserData])

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

  function renderUserNameColumn(user: User) {
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

  function renderErrorAlert() {
    return (
      <Dialog type='warning' title='Could not fetch teams'>
        {error?.error.message}
      </Dialog>
    )
  }

  function renderContent() {
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
          <Title size={2} className={pageLayoutStyles.tableTitle}>
            {teamMembersTableTitle}
          </Title>
          {teamMembersTableData.length > 0 ? (
            <Table columns={teamMembersTableHeaderColumns} data={teamMembersTableData as TableData['data']} />
          ) : (
            <Dialog type='warning' title='Ingen team funnet'>
              Du er ikke manager i noen dapla-team
            </Dialog>
          )}
        </>
      )
    }
  }

  return <PageLayout title={'Teammedlemmer'} content={renderContent()} />
}
