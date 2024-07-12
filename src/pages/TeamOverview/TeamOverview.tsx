/* eslint-disable react-hooks/exhaustive-deps */
import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Dialog, Tabs, Divider, Button } from '@statisticsnorway/ssb-component-library'

import { TabProps } from '../../@types/pageTypes'
import PageLayout from '../../components/PageLayout/PageLayout'
import Table, { TableData } from '../../components/Table/Table'
import PageSkeleton from '../../components/PageSkeleton/PageSkeleton'

import { fetchTeamOverviewData, TeamOverviewData } from '../../services/teamOverview'
import { formatDisplayName } from '../../utils/utils'
import { ApiError, fetchUserInformationFromAuthToken, isDaplaAdmin } from '../../utils/services'
import FormattedTableColumn from '../../components/FormattedTableColumn/FormattedTableColumn'
import { User } from '../../services/userProfile'
import { isAuthorizedToCreateTeam } from '../../services/createTeam'
import { Effect } from 'effect'

const MY_TEAMS_TAB = {
  title: 'Mine team',
  path: 'myTeams',
}

const ALL_TEAMS_TAB = {
  title: 'Alle teams',
  path: 'allTeams',
}

const TeamOverview = () => {
  const [activeTab, setActiveTab] = useState<TabProps | string>(MY_TEAMS_TAB)
  const [isAuthorized, setIsAuthorized] = useState(false)
  const [teamOverviewData, setTeamOverviewData] = useState<TeamOverviewData>()
  const [teamOverviewTableData, setTeamOverviewTableData] = useState<TableData['data']>()
  const [teamOverviewTableTitle, setTeamOverviewTableTitle] = useState<string>(MY_TEAMS_TAB.title)
  const [error, setError] = useState<ApiError | undefined>()
  const [loading, setLoading] = useState<boolean>(true)

  const navigate = useNavigate()

  const teamTab = (activeTab as TabProps)?.path ?? activeTab

  const prepTeamData = useCallback(
    (response: TeamOverviewData): TableData['data'] => {
      return response[teamTab].teams.map(({ uniform_name, section_name, users, managers }) => ({
        id: uniform_name,
        seksjon: section_name, // Makes section name searchable and sortable in table by including the field
        navn: <FormattedTableColumn href={`/${uniform_name}`} linkText={uniform_name} text={section_name} />,
        teammedlemmer: users.length,
        managers: managers ? managers.map((managerObj) => formatDisplayName(managerObj.display_name)).join(', ') : '',
      }))
    },
    [activeTab]
  )

  useEffect(() => {
    const fetchData = async () => {
      try {
        const tokenData = await fetchUserInformationFromAuthToken()
        if (!tokenData) return

        const response = await fetchTeamOverviewData(tokenData.email)
        setTeamOverviewData(response as TeamOverviewData)
        setTeamOverviewTableData(prepTeamData(response as TeamOverviewData))
      } catch (error) {
        setError(error as ApiError)
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [])

  useEffect(() => {
    const userProfileItem = localStorage.getItem('userProfile')
    if (!userProfileItem) return

    const user = JSON.parse(userProfileItem) as User
    if (!user) return

    Effect.promise(() => isDaplaAdmin(user.principal_name))
      .pipe(Effect.runPromise)
      .then((isDaplaAdmin: boolean) => setIsAuthorized(isAuthorizedToCreateTeam(isDaplaAdmin, user.job_title)))
  }, [])

  useEffect(() => {
    if (teamOverviewData) {
      if (teamTab === MY_TEAMS_TAB.path) {
        setTeamOverviewTableTitle(MY_TEAMS_TAB.title)
      } else {
        setTeamOverviewTableTitle(ALL_TEAMS_TAB.title)
      }
      setTeamOverviewTableData(prepTeamData(teamOverviewData)) // Update Table view on Tab onClick
    }
  }, [prepTeamData])

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
    if (loading) return <PageSkeleton />

    if (teamOverviewTableData) {
      const teamOverviewTableHeaderColumns = [
        {
          id: 'navn',
          label: 'Navn',
        },
        {
          id: 'teammedlemmer',
          label: 'Teammedlemmer',
          align: 'right',
        },
        {
          id: 'managers',
          label: 'Managers',
        },
      ]

      return (
        <>
          <Tabs
            onClick={handleTabClick}
            activeOnInit={MY_TEAMS_TAB.path}
            items={[
              {
                title: `${MY_TEAMS_TAB.title} (${teamOverviewData?.myTeams.teams.length ?? 0})`,
                path: MY_TEAMS_TAB.path,
              },
              {
                title: `${ALL_TEAMS_TAB.title} (${teamOverviewData?.allTeams.teams.length ?? 0})`,
                path: ALL_TEAMS_TAB.path,
              },
            ]}
          />
          <Divider dark />
          <Table
            title={teamOverviewTableTitle}
            columns={teamOverviewTableHeaderColumns}
            data={teamOverviewTableData as TableData['data']}
          />
        </>
      )
    }
  }

  return (
    <PageLayout
      title='Teamoversikt'
      content={renderContent()}
      button={
        <>
          {teamOverviewData && isAuthorized && <Button onClick={() => navigate('opprett-team')}>+ Opprett team</Button>}
        </>
      }
    />
  )
}

export default TeamOverview
