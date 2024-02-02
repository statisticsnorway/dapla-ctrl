import pageLayoutStyles from '../../components/PageLayout/pagelayout.module.scss'
import styles from './teamDetail.module.scss'

import { useCallback, useContext, useEffect, useState } from 'react'
import PageLayout from '../../components/PageLayout/PageLayout'
import { TeamDetailData, TeamDetailTeamResult, getTeamDetail } from '../../api/teamDetail'
import { useParams } from 'react-router-dom'
import { ErrorResponse } from '../../@types/error'
import { DaplaCtrlContext } from '../../provider/DaplaCtrlProvider'
import Table, { TableData } from '../../components/Table/Table'
import { getGroupType } from '../../utils/utils'
import { User } from '../../@types/user'
import { Text, Title, Link, Dialog } from '@statisticsnorway/ssb-component-library'
import { Skeleton } from '@mui/material'

export default function TeamDetail() {

    const { setBreadcrumbTeamDetailDisplayName } = useContext(DaplaCtrlContext)
    const [error, setError] = useState<ErrorResponse | undefined>()
    const [loadingTeamDetailData, setLoadingTeamDetailData] = useState<boolean>(true)
    const [loadingTeamData, setLoadingTeamData] = useState<boolean>(true)
    const [teamDetailData, setTeamDetailData] = useState<TeamDetailData>()
    const [teamDetailTableData, setTeamDetailTableData] = useState<TableData['data']>()

    const { teamId } = useParams<{ teamId: string }>()

    const prepTeamData = useCallback(
        (response: TeamDetailTeamResult): TableData['data'] => {
            return response.teamUsers.map((user) => ({
                id: user?.principal_name,
                navn: renderUsernameColumn(user),
                gruppe: user.groups?.map((group) => getGroupType(group.uniform_name)),
                epost: user?.principal_name
            }))
        },
        [teamDetailData]
    )

    useEffect(() => {
        if (!teamId) return
        getTeamDetail(teamId as string)
            .then((response) => {
                if ((response as ErrorResponse).error) {
                    console.log(response)
                    setError(response as ErrorResponse)
                } else {
                    console.log((response))
                    setTeamDetailData(response as TeamDetailData)
                }
            })
            .finally(() => setLoadingTeamDetailData(false))
            .catch((error) => {
                setError({ error: { message: error.message, code: '500' } })
            })

    }, [teamId])

    useEffect(() => {
        if (teamDetailData) {
            getTeamDetail(teamId as string)
                .then((response) => {
                    if ((response as ErrorResponse).error) {
                        setError(response as ErrorResponse)
                    } else {
                        setTeamDetailTableData(prepTeamData(response as TeamDetailTeamResult))
                    }
                })
                .finally(() => setLoadingTeamData(false))
                .catch((error) => {
                    setError({ error: { message: error.message, code: '500' } })
                })
        }
    }, [teamDetailData, teamId, prepTeamData])

    // required for breadcrumb
    useEffect(() => {
        if (!teamDetailData) return

        const displayName = teamDetailData.teamInfo.display_name
        teamDetailData.teamInfo.display_name = displayName
        setBreadcrumbTeamDetailDisplayName({ displayName })

    }, [teamDetailData, setBreadcrumbTeamDetailDisplayName])

    function renderUsernameColumn(user: User) {
        console.log(user)
        return (
            <>
                <span>
                    <Link href={`/teammedlemmer/${user.principal_name}`}>
                        <b>{user.display_name.split(', ').reverse().join(' ')}</b>
                    </Link>
                </span>
                {user && <Text>{user.section_name}</Text>}
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

    function renderSkeletonOnLoad() {
        return (
            <>
                <Skeleton variant='text' animation='wave' sx={{ fontSize: '5.5rem' }} width={150} />
                <Skeleton variant='rectangular' animation='wave' height={200} />
            </>
        )
    }

    function renderContent() {
        if (error) return renderErrorAlert()
        // TODO: cheesy method to exclude showing skeleton for profile information (username etc..)
        if (loadingTeamData && !loadingTeamDetailData) return renderSkeletonOnLoad()

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
                }
            ]
            return (
                <>
                    <Title size={2} className={pageLayoutStyles.tableTitle}>
                        Team
                    </Title>
                    <Table columns={teamOverviewTableHeaderColumns} data={teamDetailTableData as TableData['data']} />
                </>
            )
        }
    }

    return (
        <PageLayout
            title={teamDetailData?.teamInfo.display_name}
            content={renderContent()}
            description={
                <div className={styles.userProfileDescription}>
                    <Text medium>{"asd"}</Text>
                    <Text medium>{"asd"}</Text>
                </div>
            }
        />
    )
}
