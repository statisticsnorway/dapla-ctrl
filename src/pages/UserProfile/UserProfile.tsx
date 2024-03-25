/* eslint-disable react-hooks/exhaustive-deps */
import styles from '../../components/PageLayout/pagelayout.module.scss'

import { Dialog, Text, LeadParagraph } from '@statisticsnorway/ssb-component-library'

import Table, { TableData } from '../../components/Table/Table'
import PageLayout from '../../components/PageLayout/PageLayout'
import PageSkeleton from '../../components/PageSkeleton/PageSkeleton'

import { useCallback, useContext, useEffect, useState } from 'react'
import { DaplaCtrlContext } from '../../provider/DaplaCtrlProvider'
import { getGroupType, formatDisplayName, stripSuffixes } from '../../utils/utils'

import { getUserProfileTeamData, TeamsData } from '../../services/userProfile'

import { useParams } from 'react-router-dom'
import { Skeleton } from '@mui/material'
import { ApiError } from '../../utils/services'
import FormattedTableColumn from '../../components/FormattedTableColumn/FormattedTableColumn'

const UserProfile = () => {
  const { setBreadcrumbUserProfileDisplayName } = useContext(DaplaCtrlContext)
  const [error, setError] = useState<ApiError | undefined>()
  const [loadingTeamData, setLoadingTeamData] = useState<boolean>(true)
  const [userProfileData, setUserProfileData] = useState<TeamsData>()
  const [teamUserProfileTableData, setUserProfileTableData] = useState<TableData['data']>()
  const { principalName } = useParams()

  const prepTeamData = useCallback(
    (response: TeamsData): TableData['data'] => {
      return response.teams.map(({ uniform_name, section_name, groups, managers }) => ({
        id: uniform_name,
        seksjon: section_name, // Makes section name searchable and sortable in table by including the field
        navn: <FormattedTableColumn href={`/${uniform_name}`} linkText={uniform_name} text={section_name} />,
        gruppe: principalName
          ? groups
              ?.filter(
                (group) =>
                  group.users.some((user) => user.principal_name === principalName) &&
                  // making sure uniform name of the group and uniform name of the team are same
                  // this is to combat this issue: https://github.com/statisticsnorway/dapla-team-api-redux/issues/63
                  stripSuffixes(group.uniform_name) === uniform_name
              ) // Filter groups based on principalName presence
              .map((group) => getGroupType(group.uniform_name))
              .join(', ')
          : 'INGEN FUNNET',
        managers: managers.map((managerObj) => formatDisplayName(managerObj.display_name)).join(', '),
      }))
    },
    [principalName, userProfileData]
  )

  useEffect(() => {
    getUserProfileTeamData(principalName as string)
      .then((response) => {
        const formattedResponse = response as TeamsData
        setUserProfileTableData(prepTeamData(formattedResponse))
        setUserProfileData(formattedResponse)

        const displayName = formatDisplayName(formattedResponse.user.display_name)
        setBreadcrumbUserProfileDisplayName({ displayName })
      })
      .finally(() => setLoadingTeamData(false))
      .catch((error) => {
        setError(error as ApiError)
      })
  }, [principalName, setBreadcrumbUserProfileDisplayName])

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

    if (teamUserProfileTableData) {
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
          id: 'managers',
          label: 'Managers',
        },
      ]
      return (
        <>
          <LeadParagraph className={styles.description}>
            <Text medium>{userProfileData?.user.section_name}</Text>
            <Text medium>{userProfileData?.user.principal_name}</Text>
          </LeadParagraph>
          <Table
            title='Team'
            columns={teamOverviewTableHeaderColumns}
            data={teamUserProfileTableData as TableData['data']}
          />
        </>
      )
    }
  }

  return (
    <PageLayout
      title={
        !loadingTeamData && userProfileData ? (
          (formatDisplayName(userProfileData?.user.display_name) as string)
        ) : (
          <Skeleton variant='rectangular' animation='wave' width={350} height={90} />
        )
      }
      content={renderContent()}
    />
  )
}

export default UserProfile
