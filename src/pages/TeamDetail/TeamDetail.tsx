/* eslint-disable react-hooks/exhaustive-deps */
import styles from './teamDetail.module.scss'

import { useCallback, useContext, useEffect, useState } from 'react'
import PageLayout from '../../components/PageLayout/PageLayout'
import { TeamDetailData, getTeamDetail } from '../../services/teamDetail'
import { useParams } from 'react-router-dom'
import { ApiError } from '../../utils/services'

import { DaplaCtrlContext } from '../../provider/DaplaCtrlProvider'
import Table, { TableData } from '../../components/Table/Table'
import { formatDisplayName, getGroupType } from '../../utils/utils'
import { User } from '../../services/teamDetail'
import { Text, Link, Dialog, LeadParagraph } from '@statisticsnorway/ssb-component-library'
import PageSkeleton from '../../components/PageSkeleton/PageSkeleton'
import { Skeleton } from '@mui/material'

const TeamDetail = () => {
  const { setBreadcrumbTeamDetailDisplayName } = useContext(DaplaCtrlContext)
  const [error, setError] = useState<ApiError | undefined>()
  const [loadingTeamData, setLoadingTeamData] = useState<boolean>(true)
  const [teamDetailData, setTeamDetailData] = useState<TeamDetailData>()
  const [teamDetailTableData, setTeamDetailTableData] = useState<TableData['data']>()

  const { teamId } = useParams<{ teamId: string }>()

  const prepTeamData = useCallback((response: TeamDetailData): TableData['data'] => {
    if (!response.team.users) {
      return []
    }

    return response.team.users.map((user) => {
      // Makes data in username column searchable and sortable in table by including these fields
      const usernameColumn = {
        user: user.display_name,
        seksjon: user.section_name,
      }

      return {
        id: user?.principal_name,
        ...usernameColumn,
        navn: renderUsernameColumn(user),
        gruppe: user.groups
          ?.filter((group) => group.uniform_name.startsWith(response.team.uniform_name))
          .map((group) => getGroupType(group.uniform_name))
          .join(', '),
        epost: user?.principal_name,
      }
    })
  }, [])

  useEffect(() => {
    if (!teamId) return
    getTeamDetail(teamId)
      .then((response) => {
        const formattedResponse = response as TeamDetailData
        setTeamDetailData(formattedResponse)

        const displayName = formatDisplayName(formattedResponse.team.display_name)
        setBreadcrumbTeamDetailDisplayName({ displayName })
      })
      .catch((error) => {
        setError(error as ApiError)
      })
  }, [teamId, setBreadcrumbTeamDetailDisplayName])

  useEffect(() => {
    getTeamDetail(teamId as string)
      .then((response) => {
        setTeamDetailTableData(prepTeamData(response as TeamDetailData))
      })
      .finally(() => setLoadingTeamData(false))
      .catch((error) => {
        setError(error as ApiError)
      })
  }, [prepTeamData])

  const renderUsernameColumn = (user: User) => {
    return (
      <>
        <span>
          <Link href={`/teammedlemmer/${user.principal_name}`}>
            <b>{formatDisplayName(user.display_name)}</b>
          </Link>
        </span>
        {user && <Text>{user.section_name ? user.section_name : 'Mangler seksjon'}</Text>}
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
            <Text medium className={styles.uniformName}>
              {teamDetailData ? teamDetailData.team.uniform_name : ''}
            </Text>
            <Text medium>
              {teamDetailData ? formatDisplayName(teamDetailData.team.manager?.display_name ?? '') : ''}
            </Text>
            <Text medium>{teamDetailData ? teamDetailData.team.section_name : ''}</Text>
          </LeadParagraph>
          <Table
            title='Teammedlemmer'
            columns={teamOverviewTableHeaderColumns}
            data={teamDetailTableData as TableData['data']}
          />
        </>
      )
    }
  }

  return (
    <PageLayout
      title={
        !loadingTeamData && teamDetailData ? (
          teamDetailData.team.display_name
        ) : (
          <Skeleton variant='rectangular' animation='wave' width={350} height={90} />
        )
      }
      content={renderContent()}
    />
  )
}

export default TeamDetail
