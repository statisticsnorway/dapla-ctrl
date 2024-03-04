/* eslint-disable react-hooks/exhaustive-deps */
import styles from './teamDetail.module.scss'

import { useCallback, useContext, useEffect, useState } from 'react'
import PageLayout from '../../components/PageLayout/PageLayout'
import { TeamDetailData, getTeamDetail, Team, SharedBucket, SharedBuckets } from '../../services/teamDetail'
import { useParams } from 'react-router-dom'
import { ApiError } from '../../utils/services'

import { DaplaCtrlContext } from '../../provider/DaplaCtrlProvider'
import Table, { TableData } from '../../components/Table/Table'
import { formatDisplayName, getGroupType } from '../../utils/utils'
import { User } from '../../services/teamDetail'
import { Text, Link, Dialog, LeadParagraph, Divider, Tabs } from '@statisticsnorway/ssb-component-library'
import PageSkeleton from '../../components/PageSkeleton/PageSkeleton'
import { Skeleton, TabProps } from '@mui/material'

const TeamDetail = () => {
  const defaultActiveTab = {
    title: 'Teammedlemmer',
    path: 'team',
  }

  const [activeTab, setActiveTab] = useState<TabProps | string>(defaultActiveTab)

  const { setBreadcrumbTeamDetailDisplayName } = useContext(DaplaCtrlContext)
  const [error, setError] = useState<ApiError | undefined>()
  const [loadingTeamData, setLoadingTeamData] = useState<boolean>(true)
  const [teamDetailData, setTeamDetailData] = useState<TeamDetailData>()
  const [teamDetailTableTitle, setTeamDetailTableTitle] = useState<string>(defaultActiveTab.title)
  const [teamDetailTableData, setTeamDetailTableData] = useState<TableData['data']>()

  const { teamId } = useParams<{ teamId: string }>()

  const prepTeamData = useCallback(
    (response: TeamDetailData): TableData['data'] => {
      const teamDetailTab = (activeTab as TabProps)?.path ?? activeTab
      if (teamDetailTab === 'sharedBuckets') {
        if (!response['sharedBuckets'].items) return []
        return response['sharedBuckets'].items.map(({ short_name, bucket_name, metrics }) => {
          return {
            id: short_name,
            navn: renderBucketNameColumn({ short_name, bucket_name }),
            tilgang: `${metrics[0].teams_count} team`,
            delte_data: metrics[0].groups_count,
            antall_personer: metrics[0].users_count,
          }
        })
      } else {
        if (!response['team'].users) return []

        return response['team'].users.map((user) => {
          return {
            id: formatDisplayName(user.display_name),
            navn: renderUsernameColumn(user),
            seksjon: user.section_name, // Makes section name searchable and sortable in table by including the field
            gruppe: user.groups
              ?.filter((group) => group.uniform_name.startsWith(response.team.uniform_name))
              .map((group) => getGroupType(group.uniform_name))
              .join(', '),
            epost: user?.principal_name,
          }
        })
      }
    },
    [activeTab]
  )

  useEffect(() => {
    if (!teamId) return
    getTeamDetail(teamId)
      .then((response) => {
        const formattedResponse = response as TeamDetailData
        setTeamDetailData(formattedResponse)

        const displayName = formatDisplayName((formattedResponse.team as Team).display_name)
        setBreadcrumbTeamDetailDisplayName({ displayName })
      })
      .catch((error) => {
        setError(error as ApiError)
      })
  }, [])

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

  const handleTabClick = (tab: string) => {
    setActiveTab(tab)
    if (tab === defaultActiveTab.path) {
      setTeamDetailTableTitle(defaultActiveTab.title)
    } else {
      setTeamDetailTableTitle('Delte data')
    }
  }

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

  const renderBucketNameColumn = ({ short_name, bucket_name }: SharedBucket) => {
    return (
      <>
        <span>
          <Link href={`/${teamId}/${short_name}`}>
            <b>{short_name}</b>
          </Link>
        </span>
        <Text>{bucket_name}</Text>
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
      const teamOverviewTableHeaderColumns =
        activeTab === 'sharedBuckets'
          ? [
              {
                id: 'navn',
                label: 'Navn',
              },
              {
                id: 'tilgang',
                label: 'Tilgang',
              },
              { id: 'delte_data', label: 'Delte data' },
              { id: 'antall_personer', label: 'Antall personer' },
            ]
          : [
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
              {teamDetailData ? (teamDetailData.team as Team).uniform_name : ''}
            </Text>
            <Text medium>
              {teamDetailData ? formatDisplayName((teamDetailData.team as Team).manager?.display_name ?? '') : ''}
            </Text>
            <Text medium>{teamDetailData ? (teamDetailData.team as Team).section_name : ''}</Text>
          </LeadParagraph>
          <Tabs
            onClick={handleTabClick}
            activeOnInit={defaultActiveTab.path}
            items={[
              {
                title: `${defaultActiveTab.title} (${(teamDetailData?.team as Team).users?.length ?? 0})`,
                path: defaultActiveTab.path,
              },
              {
                title: `Delte data (${(teamDetailData?.sharedBuckets as SharedBuckets).items?.length ?? 0})`,
                path: 'sharedBuckets',
              },
            ]}
          />
          <Divider dark />
          <Table
            title={teamDetailTableTitle}
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
          (teamDetailData.team as Team).display_name
        ) : (
          <Skeleton variant='rectangular' animation='wave' width={350} height={90} />
        )
      }
      content={renderContent()}
    />
  )
}

export default TeamDetail
