/* eslint-disable react-hooks/exhaustive-deps */
import styles from '../../components/PageLayout/pagelayout.module.scss'

import PageLayout from '../../components/PageLayout/PageLayout'
import PageSkeleton from '../../components/PageSkeleton/PageSkeleton'
import Table, { TableData } from '../../components/Table/Table'
import { ApiError } from '../../utils/services'

import { DaplaCtrlContext } from '../../provider/DaplaCtrlProvider'

import { useState, useEffect, useContext } from 'react'
import { useParams } from 'react-router-dom'
import { Dialog, LeadParagraph, Text } from '@statisticsnorway/ssb-component-library'
import { Team, SharedBucket, SharedBucketDetail, getSharedBucketsDetailData } from '../../services/sharedBucketsDetail'
import FormattedTableColumn from '../../components/FormattedTableColumn'

const SharedBucketDetail = () => {
  const { setBreadcrumbTeamDetailDisplayName, setBreadcrumbBucketDetailDisplayName } = useContext(DaplaCtrlContext)
  const [error, setError] = useState<ApiError | undefined>()
  const [loadingSharedBucketsData, setLoadingSharedBucketsData] = useState<boolean>(true)
  const [sharedBucketData, setSharedBucketData] = useState<SharedBucketDetail>()
  const [sharedBucketTableData, setSharedBucketTableData] = useState<TableData['data']>()

  const { teamId } = useParams<{ teamId: string }>()
  const { shortName } = useParams<{ shortName: string }>()

  const prepSharedBucketTableData = (response: SharedBucketDetail): TableData['data'] => {
    return (response['sharedBucket'] as SharedBucket).teams.map(({ display_name, uniform_name, section_name }) => {
      return {
        id: display_name ?? '',
        team: <FormattedTableColumn href={`/${uniform_name}`} linkText={display_name} text={section_name} />,
      }
    })
  }

  useEffect(() => {
    if (!teamId && !shortName) return
    getSharedBucketsDetailData(teamId as string, shortName as string)
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
          id: 'team',
          label: 'Team',
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
