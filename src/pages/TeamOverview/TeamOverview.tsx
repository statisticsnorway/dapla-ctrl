/* eslint-disable react-hooks/exhaustive-deps */
import { useCallback, useEffect, useState } from 'react'
import { Dialog, Text, Link, Tabs, Divider } from '@statisticsnorway/ssb-component-library'

import { TabProps } from '../../@types/pageTypes'
import PageLayout from '../../components/PageLayout/PageLayout'
import Table, { TableData } from '../../components/Table/Table'
import PageSkeleton from '../../components/PageSkeleton/PageSkeleton'

import { ErrorResponse } from '../../@types/error'
import { Team } from '../../@types/team'

import { getTeamOverview, TeamOverviewData } from '../../services/teamOverview'
import { formatDisplayName } from '../../utils/utils'

const TeamOverview = () => {
  const defaultActiveTab = {
    title: 'Mine team',
    path: 'myTeams',
  }

  const [activeTab, setActiveTab] = useState<TabProps | string>(defaultActiveTab)
  const [teamOverviewData, setTeamOverviewData] = useState<TeamOverviewData>()
  const [teamOverviewTableData, setTeamOverviewTableData] = useState<TableData['data']>()
  const [teamOverviewTableTitle, setTeamOverviewTableTitle] = useState<string>(defaultActiveTab.title)
  const [error, setError] = useState<ErrorResponse | undefined>()
  const [loading, setLoading] = useState<boolean>(true)

  const prepTeamData = useCallback(
    (response: TeamOverviewData): TableData['data'] => {
      const team = (activeTab as TabProps)?.path ?? activeTab

      return response[team].teams.map((team) => ({
        id: team.uniform_name,
        seksjon: team.section_name, // Makes section name searchable and sortable in table by including the field
        navn: renderTeamNameColumn(team),
        teammedlemmer: team.team_user_count,
        ansvarlig: formatDisplayName(team.manager.display_name),
      }))
    },
    [activeTab]
  )

  useEffect(() => {
    getTeamOverview()
      .then((response) => {
        if ((response as ErrorResponse).error) {
          setError(response as ErrorResponse)
        } else {
          setTeamOverviewData(response as TeamOverviewData)
          setTeamOverviewTableData(prepTeamData(response as TeamOverviewData))
        }
      })
      .finally(() => setLoading(false))
      .catch((error) => {
        setError(error.toString())
      })
  }, [])

  useEffect(() => {
    if (teamOverviewData) {
      setTeamOverviewTableData(prepTeamData(teamOverviewData)) // Update Table view on Tab onClick
    }
  }, [prepTeamData])

  const handleTabClick = (tab: string) => {
    setActiveTab(tab)
    if (tab === 'myTeams') {
      setTeamOverviewTableTitle('Mine team')
    } else {
      setTeamOverviewTableTitle('Alle teams')
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
        {error?.error.message}
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
            activeOnInit={defaultActiveTab.path}
            items={[
              { title: `Mine team (${teamOverviewData?.myTeams.count ?? 0})`, path: 'myTeams' },
              { title: `Alle team (${teamOverviewData?.allTeams.count ?? 0})`, path: 'allTeams' },
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
