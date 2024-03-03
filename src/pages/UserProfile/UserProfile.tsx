import styles from './userprofile.module.scss'

import { Dialog, Text, Link, LeadParagraph } from '@statisticsnorway/ssb-component-library'

import Table, { TableData } from '../../components/Table/Table'
import PageLayout from '../../components/PageLayout/PageLayout'
import PageSkeleton from '../../components/PageSkeleton/PageSkeleton'

import { useCallback, useContext, useEffect, useState } from 'react'
import { DaplaCtrlContext } from '../../provider/DaplaCtrlProvider'
import { getGroupType, formatDisplayName } from '../../utils/utils'

import { getUserProfileTeamData, TeamsData, Team, UserProfileTeamData } from '../../services/userProfile'

import { useParams } from 'react-router-dom'
import { Skeleton } from '@mui/material'
import { ApiError } from '../../utils/services'

const UserProfile = () => {
  const { setBreadcrumbUserProfileDisplayName } = useContext(DaplaCtrlContext)
  const [error, setError] = useState<ApiError | undefined>()
  const [loadingTeamData, setLoadingTeamData] = useState<boolean>(true)
  const [userProfileData, setUserProfileData] = useState<TeamsData>()
  const [teamUserProfileTableData, setUserProfileTableData] = useState<TableData['data']>()
  const { principalName } = useParams()

  const prepTeamData = useCallback(
    (response: TeamsData): TableData['data'] => {
      return response.teams.map((team) => ({
        id: team.uniform_name,
        seksjon: team.section_name, // Makes section name searchable and sortable in table by including the field
        navn: renderTeamNameColumn(team),
        gruppe: principalName ? team.groups
        ?.filter(group => group.users.some(user => user.principal_name === principalName)) // Filter groups based on principalName presence
        .map(group => getGroupType(group.uniform_name))
        .join(', ') : "INGEN FUNNET",
        ansvarlig: formatDisplayName(team.manager.display_name),
      }))
    },
    [principalName]
  )

  useEffect(() => {
      getUserProfileTeamData(principalName as string)
        .then((response) => {
          const formattedResponse = response as TeamsData;
          setUserProfileTableData(prepTeamData(formattedResponse))
          setUserProfileData((formattedResponse))

          const displayName = formatDisplayName(formattedResponse.user.display_name);
          setBreadcrumbUserProfileDisplayName({ displayName });
        })
        .finally(() => setLoadingTeamData(false))
        .catch((error) => {
          setError(error as ApiError)
        })
    
  }, [principalName, prepTeamData])

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
          id: 'ansvarlig',
          label: 'Ansvarlig',
        },
      ]
      return (
        <>
          <LeadParagraph className={styles.userProfileDescription}>
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
