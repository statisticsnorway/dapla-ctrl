import styles from './userprofile.module.scss'

import { Dialog, Text, Link, LeadParagraph } from '@statisticsnorway/ssb-component-library'

import Table, { TableData } from '../../components/Table/Table'
import PageLayout from '../../components/PageLayout/PageLayout'
import PageSkeleton from '../../components/PageSkeleton/PageSkeleton'

import { useCallback, useContext, useEffect, useState } from 'react'
import { DaplaCtrlContext } from '../../provider/DaplaCtrlProvider'
import { getGroupType, formatDisplayName } from '../../utils/utils'

import { getUserProfile, getUserTeamsWithGroups, UserProfileTeamResult } from '../../services/userProfile'

import { User } from '../../@types/user'
import { Team } from '../../@types/team'

import { useParams } from 'react-router-dom'
import { ErrorResponse } from '../../@types/error'
import { Skeleton } from '@mui/material'

export default function UserProfile() {
  const { setBreadcrumbUserProfileDisplayName } = useContext(DaplaCtrlContext)
  const [error, setError] = useState<ErrorResponse | undefined>()
  const [loadingTeamData, setLoadingTeamData] = useState<boolean>(true)
  const [userProfileData, setUserProfileData] = useState<User>()
  const [teamUserProfileTableData, setUserProfileTableData] = useState<TableData['data']>()
  const { principalName } = useParams()

  const prepTeamData = useCallback(
    (response: UserProfileTeamResult): TableData['data'] => {
      return response.teams.map((team) => ({
        id: team.uniform_name,
        seksjon: team.section_name, // Makes section name searchable and sortable in table by including the field
        navn: renderTeamNameColumn(team),
        gruppe: team.groups?.map((group) => getGroupType(group)).join(', '),
        epost: userProfileData?.principal_name,
        ansvarlig: formatDisplayName(team.manager.display_name),
      }))
    },
    [userProfileData]
  )

  useEffect(() => {
    getUserProfile(principalName as string)
      .then((response) => {
        if ((response as ErrorResponse).error) {
          setError(response as ErrorResponse)
        } else {
          setUserProfileData(response as User)
        }
      })
      .catch((error) => {
        setError({ error: { message: error.message, code: '500' } })
      })
  }, [principalName])

  useEffect(() => {
    if (userProfileData) {
      getUserTeamsWithGroups(principalName as string)
        .then((response) => {
          if ((response as ErrorResponse).error) {
            setError(response as ErrorResponse)
          } else {
            setUserProfileTableData(prepTeamData(response as UserProfileTeamResult))
          }
        })
        .finally(() => setLoadingTeamData(false))
        .catch((error) => {
          setError({ error: { message: error.message, code: '500' } })
        })
    }
  }, [userProfileData, principalName, prepTeamData])

  // required for breadcrumb
  useEffect(() => {
    if (userProfileData) {
      const displayName = formatDisplayName(userProfileData.display_name)
      userProfileData.display_name = displayName
      setBreadcrumbUserProfileDisplayName({ displayName })
    }
  }, [userProfileData, setBreadcrumbUserProfileDisplayName])

  function renderTeamNameColumn(team: Team) {
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
          id: 'epost',
          label: 'Epost ?',
        },
        {
          id: 'ansvarlig',
          label: 'Ansvarlig',
        },
      ]
      return (
        <>
          <LeadParagraph className={styles.userProfileDescription}>
            <Text medium>{userProfileData?.section_name}</Text>
            <Text medium>{userProfileData?.principal_name}</Text>
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
          (userProfileData?.display_name as string)
        ) : (
          <Skeleton variant='rectangular' animation='wave' width={350} height={90} />
        )
      }
      content={renderContent()}
    />
  )
}
