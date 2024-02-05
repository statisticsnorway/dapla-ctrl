import pageLayoutStyles from '../../components/PageLayout/pagelayout.module.scss'
import styles from './teamDetail.module.scss'

import { useCallback, useContext, useEffect, useState } from 'react'
import PageLayout from '../../components/PageLayout/PageLayout'
import { TeamDetailData, getTeamDetail } from '../../services/teamDetail'
import { useParams } from 'react-router-dom'
import { ErrorResponse } from '../../@types/error'
import { DaplaCtrlContext } from '../../provider/DaplaCtrlProvider'
import Table, { TableData } from '../../components/Table/Table'
import { getGroupType } from '../../utils/utils'
import { User } from '../../@types/user'
import { Text, Title, Link, Dialog, LeadParagraph } from '@statisticsnorway/ssb-component-library'
import PageSkeleton from '../../components/PageSkeleton/PageSkeleton'

export default function TeamDetail() {
  const { setBreadcrumbTeamDetailDisplayName } = useContext(DaplaCtrlContext)
  const [error, setError] = useState<ErrorResponse | undefined>()
  const [loadingTeamData, setLoadingTeamData] = useState<boolean>(true)
  const [teamDetailData, setTeamDetailData] = useState<TeamDetailData>()
  const [teamDetailTableData, setTeamDetailTableData] = useState<TableData['data']>()

  const { teamId } = useParams<{ teamId: string }>()

  const prepTeamData = useCallback((response: TeamDetailData): TableData['data'] => {
    return response['teamUsers'].teamUsers.map((user) => ({
      id: user?.principal_name,
      navn: renderUsernameColumn(user),
      gruppe: user.groups?.map((group) => getGroupType(group.uniform_name)).join(', '),
      epost: user?.principal_name,
    }))
  }, [])

  useEffect(() => {
    if (!teamId) return
    getTeamDetail(teamId as string)
      .then((response) => {
        if ((response as ErrorResponse).error) {
          setError(response as ErrorResponse)
        } else {
          setTeamDetailData(response as TeamDetailData)
        }
      })
      .catch((error) => {
        setError({ error: { message: error.message, code: '500' } })
      })
  }, [teamId])

  useEffect(() => {
    getTeamDetail(teamId as string)
      .then((response) => {
        if ((response as ErrorResponse).error) {
          setError(response as ErrorResponse)
        } else {
          setTeamDetailTableData(prepTeamData(response as TeamDetailData))
        }
      })
      .finally(() => setLoadingTeamData(false))
      .catch((error) => {
        setError({ error: { message: error.message, code: '500' } })
      })
  }, [teamId, prepTeamData])

  // required for breadcrumb
  useEffect(() => {
    if (!teamDetailData) return

    const displayName = teamDetailData['teamUsers'].teamInfo.display_name
    teamDetailData['teamUsers'].teamInfo.display_name = displayName
    setBreadcrumbTeamDetailDisplayName({ displayName })
  }, [teamDetailData, setBreadcrumbTeamDetailDisplayName])

  function renderUsernameColumn(user: User) {
    return (
      <>
        <span>
          <Link href={`/teammedlemmer/${user.principal_name.split('@')[0]}`}>
            <b>{user.display_name.split(', ').reverse().join(' ')}</b>
          </Link>
        </span>
        {user && <Text>{user.section_name ? user.section_name : 'Mangler seksjon'}</Text>}
      </>
    )
  }

  function renderErrorAlert() {
    return (
      <Dialog type='warning' title='Could not fetch data'>
        {error?.error.message}
      </Dialog>
    )
  }

  function renderContent() {
    if (error) return renderErrorAlert()
    if (loadingTeamData) return <PageSkeleton hasDescription hasTab={false} /> // TODO: Remove hasTab prop after tabs are implemented

    if (teamDetailTableData) {
      const teamOverviewTableHeaderColumns = [
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
          label: 'Epost ?',
        },
      ]
      return (
        <>
          <LeadParagraph className={styles.userProfileDescription}>
            <Text medium>{teamDetailData ? teamDetailData['teamUsers'].teamInfo.uniform_name : ''}</Text>
            <Text medium>
              {teamDetailData
                ? teamDetailData['teamUsers'].teamInfo.manager.display_name.split(', ').reverse().join(' ')
                : ''}
            </Text>
            <Text medium>{teamDetailData ? teamDetailData['teamUsers'].teamInfo.section_name : ''}</Text>
          </LeadParagraph>
          <Title size={2} className={pageLayoutStyles.tableTitle}>
            Teammedlemmer
          </Title>
          <Table columns={teamOverviewTableHeaderColumns} data={teamDetailTableData as TableData['data']} />
        </>
      )
    }
  }

  return (
    <PageLayout
      title={teamDetailData ? teamDetailData['teamUsers'].teamInfo.display_name : ''}
      content={renderContent()}
    />
  )
}
