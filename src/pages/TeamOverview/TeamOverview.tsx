/* eslint-disable react-hooks/exhaustive-deps */
import { useCallback, useEffect, useState } from 'react'
import { Dialog, Text, Link, Tabs, Divider } from '@statisticsnorway/ssb-component-library'

import { TabProps } from '../../@types/pageTypes'
import PageLayout from '../../components/PageLayout/PageLayout'
import Table, { TableData } from '../../components/Table/Table'
import PageSkeleton from '../../components/PageSkeleton/PageSkeleton'

import { fetchTeamOverviewData, TeamOverviewData, Team } from '../../services/teamOverview'
import { formatDisplayName } from '../../utils/utils'
import { ApiError } from '../../utils/services'

const MY_TEAMS_TAB = {
  title: 'Mine team',
  path: 'myTeams',
}

const ALL_TEAMS_TAB = {
  title: 'Alle teams',
  path: 'allTeams',
}

const TeamOverview = () => {
  const accessToken = localStorage.getItem('access_token') || ''
  const jwt = JSON.parse(atob(accessToken.split('.')[1]))

  const [activeTab, setActiveTab] = useState<TabProps | string>(MY_TEAMS_TAB)
  const [teamOverviewData, setTeamOverviewData] = useState<TeamOverviewData>()
  const [teamOverviewTableData, setTeamOverviewTableData] = useState<TableData['data']>()
  const [teamOverviewTableTitle, setTeamOverviewTableTitle] = useState<string>(MY_TEAMS_TAB.title)
  const [error, setError] = useState<ApiError | undefined>()
  const [loading, setLoading] = useState<boolean>(true)

  const prepTeamData = useCallback(
    (response: TeamOverviewData): TableData['data'] => {
      const teamTab = (activeTab as TabProps)?.path ?? activeTab
      return response[teamTab].teams.map((team) => ({
        id: team.uniform_name,
        seksjon: team.section_name, // Makes section name searchable and sortable in table by including the field
        navn: renderTeamNameColumn(team),
        teammedlemmer: team.users.length,
        ansvarlig: formatDisplayName(team.manager.display_name),
      }))
    },
    [activeTab]
  )

  useEffect(() => {
    if (!jwt) return
    fetchTeamOverviewData(jwt.email)
      .then((response) => {
        setTeamOverviewData(response as TeamOverviewData)
        setTeamOverviewTableData(prepTeamData(response as TeamOverviewData))
      })
      .finally(() => setLoading(false))
      .catch((error) => {
        setError(error as ApiError)
      })
  }, [])

  useEffect(() => {
    if (teamOverviewData) {
      setTeamOverviewTableData(prepTeamData(teamOverviewData)) // Update Table view on Tab onClick
    }
  }, [prepTeamData])

  const handleTabClick = (tab: string) => {
    setActiveTab(tab)
    if (tab === MY_TEAMS_TAB.path) {
      setTeamOverviewTableTitle(MY_TEAMS_TAB.title)
    } else {
      setTeamOverviewTableTitle(ALL_TEAMS_TAB.title)
    }
  }

  const renderTeamNameColumn = (team: Team) => {
    return (
      <>
        <span>
          <Link href={`/${team.uniform_name}`}>
            <b>{team.uniform_name}</b>
          </Link>
        </span>
        {team.section_name && <Text>{team.section_name}</Text>}
      </>
    )
  }

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
        },
        {
          id: 'ansvarlig',
          label: 'Ansvarlig',
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

  return <PageLayout title='Teamoversikt' content={renderContent()} />
}

export default TeamOverview
