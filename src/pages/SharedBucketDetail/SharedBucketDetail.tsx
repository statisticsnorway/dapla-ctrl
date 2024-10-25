/* eslint-disable react-hooks/exhaustive-deps */
import styles from '../../components/PageLayout/pagelayout.module.scss'

import PageLayout from '../../components/PageLayout/PageLayout'
import PageSkeleton from '../../components/PageSkeleton/PageSkeleton'
import FormattedTableColumn from '../../components/FormattedTableColumn/FormattedTableColumn'
import Table, { TableData } from '../../components/Table/Table'
import { ApiError } from '../../utils/services'
import {
  Team,
  SharedBucket,
  SharedBucketDetail as SharedBucketDetailType,
  getSharedBucketDetailData,
} from '../../services/sharedBucketDetail'

import { DaplaCtrlContext } from '../../provider/DaplaCtrlProvider'

import { useState, useEffect, useContext } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Dialog, LeadParagraph, Text, Link } from '@statisticsnorway/ssb-component-library'
import { formatDisplayName, getGroupType, getTeamFromGroup } from '../../utils/utils'

const SharedBucketDetail = () => {
  const { setBreadcrumbTeamDetailDisplayName, setBreadcrumbBucketDetailDisplayName } = useContext(DaplaCtrlContext)
  const [error, setError] = useState<ApiError | undefined>()
  const [loadingSharedBucketsData, setLoadingSharedBucketsData] = useState<boolean>(true)
  const [sharedBucketData, setSharedBucketData] = useState<SharedBucketDetailType>()
  const [sharedBucketTableData, setSharedBucketTableData] = useState<TableData['data']>()

  const { teamId } = useParams<{ teamId: string }>()
  const { shortName } = useParams<{ shortName: string }>()
  const navigate = useNavigate()

  const prepSharedBucketTableData = (response: SharedBucketDetailType): TableData['data'] => {
    const usersMap: {
      [principalName: string]: {
        id: string
        navn: JSX.Element
        seksjon: string
        principal_name: string
        gruppe: string[]
        team: JSX.Element
        team_navn: string
      }
    } = {}

    const allTeams = response['allTeams'] as Team[]
    ;((response['sharedBucket'] as SharedBucket).groups ?? []).forEach(({ uniform_name, users }) => {
      ;(users ?? []).forEach((user) => {
        const team_name = getTeamFromGroup(allTeams, uniform_name)
        const group_type = getGroupType(team_name, uniform_name)
        const key = `${user.principal_name}-${team_name}`
        if (!usersMap[key]) {
          usersMap[key] = {
            id: formatDisplayName(user.display_name),
            navn: (
              <FormattedTableColumn
                href={`/teammedlemmer/${user.principal_name}`}
                linkText={formatDisplayName(user.display_name)}
                text={user.section_name}
              />
            ),
            seksjon: user.section_name,
            principal_name: user.principal_name,
            team_navn: team_name,
            gruppe: [group_type],
            team: (
              <span>
                <Link href={`/${team_name}`}>{team_name}</Link>
              </span>
            ),
          }
        } else {
          usersMap[key].gruppe.push(getGroupType(team_name, uniform_name))
        }
      })
    })
    return Object.values(usersMap).map((user) => ({
      ...user,
      gruppe: user.gruppe.join(', '),
    }))
  }

  useEffect(() => {
    if (!teamId && !shortName) return
    getSharedBucketDetailData(teamId as string, shortName as string)
      .then((response) => {
        setSharedBucketData(response)
        setSharedBucketTableData(prepSharedBucketTableData(response))

        const teamDisplayName = (response.team as Team).display_name as string
        setBreadcrumbTeamDetailDisplayName({ displayName: teamDisplayName })

        const bucketDisplayName = (response.sharedBucket as SharedBucket).bucket_name
        setBreadcrumbBucketDetailDisplayName({ displayName: bucketDisplayName })
      })
      .finally(() => setLoadingSharedBucketsData(false))
      .catch((error) => {
        if (error?.code === 404) return navigate('/not-found')
        setError(error as ApiError)
      })
  }, [])

  const renderErrorAlert = () => {
    return (
      <Dialog type='warning' title='Could not fetch shared buckets detail'>
        {`${error?.code} - ${error?.message}`}
      </Dialog>
    )
  }

  const renderContent = () => {
    if (error) return renderErrorAlert()
    if (loadingSharedBucketsData) return <PageSkeleton hasDescription hasTab={false} />

    if (sharedBucketData) {
      const sharedBucketsTableHeaderColumns = [
        {
          id: 'navn',
          label: 'Navn',
        },
        {
          id: 'team',
          label: 'Team',
        },
        {
          id: 'gruppe',
          label: 'Gruppe',
        },
      ]

      return (
        <>
          <LeadParagraph className={styles.description}>
            <Text medium className={styles.descriptionSpacing}>
              {(sharedBucketData.sharedBucket as SharedBucket).bucket_name}
            </Text>
            <Text medium>{(sharedBucketData.team as Team).display_name}</Text>
            <Text medium>{(sharedBucketData.team as Team).section_name}</Text>
          </LeadParagraph>
          <Table
            title='Tilganger'
            columns={sharedBucketsTableHeaderColumns}
            data={sharedBucketTableData as TableData['data']}
          />
        </>
      )
    }
  }

  return (
    <PageLayout
      title={sharedBucketData ? (sharedBucketData.sharedBucket as SharedBucket).short_name : shortName}
      content={renderContent()}
    />
  )
}

export default SharedBucketDetail
