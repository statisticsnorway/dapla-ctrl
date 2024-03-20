/* eslint-disable react-hooks/exhaustive-deps */
import styles from '../../components/PageLayout/pagelayout.module.scss'

import PageLayout from '../../components/PageLayout/PageLayout'
import PageSkeleton from '../../components/PageSkeleton/PageSkeleton'
import FormattedTableColumn from '../../components/FormattedTableColumn'
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
import { useParams } from 'react-router-dom'
import { Dialog, LeadParagraph, Text } from '@statisticsnorway/ssb-component-library'
import { formatDisplayName, getGroupType, stripSuffixes } from '../../utils/utils'

const SharedBucketDetail = () => {
  const { setBreadcrumbTeamDetailDisplayName, setBreadcrumbBucketDetailDisplayName } = useContext(DaplaCtrlContext)
  const [error, setError] = useState<ApiError | undefined>()
  const [loadingSharedBucketsData, setLoadingSharedBucketsData] = useState<boolean>(true)
  const [sharedBucketData, setSharedBucketData] = useState<SharedBucketDetailType>()
  const [sharedBucketTableData, setSharedBucketTableData] = useState<TableData['data']>()

  const { teamId } = useParams<{ teamId: string }>()
  const { shortName } = useParams<{ shortName: string }>()

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

    ;(response['sharedBucket'] as SharedBucket).groups.forEach(({ uniform_name, users }) => {
      ;(users ?? []).forEach((user) => {
        if (!usersMap[user.principal_name]) {
          usersMap[user.principal_name] = {
            id: user.display_name,
            navn: (
              <FormattedTableColumn
                href={`/teammedlemmer/${user.principal_name}`}
                linkText={formatDisplayName(user.display_name)}
                text={user.section_name}
              />
            ),
            seksjon: user.section_name,
            principal_name: user.principal_name,
            team_navn: stripSuffixes(uniform_name),
            gruppe: [getGroupType(uniform_name)],
            team: (
              <FormattedTableColumn href={`/${stripSuffixes(uniform_name)}`} linkText={stripSuffixes(uniform_name)} />
            ),
          }
        } else {
          usersMap[user.principal_name].gruppe.push(getGroupType(uniform_name))
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
