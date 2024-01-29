import pageLayoutStyles from '../../components/PageLayout/pagelayout.module.scss'

import { useEffect, useState } from "react"
import PageLayout from '../../components/PageLayout/PageLayout'
import { TabProps } from '../../@types/pageTypes'
import Table, { TableData } from '../../components/Table/Table'
import { getTeamOverview, TeamOverviewData, TeamOverviewError, Team } from "../../api/teamOverview"
import { Dialog, Title, Text, Link, Tabs, Divider } from "@statisticsnorway/ssb-component-library"
import Skeleton from '@mui/material/Skeleton';

export default function TeamOverview() {
    const defaultActiveTab = {
        title: 'Mine team',
        path: 'myTeams'
    }

    const [activeTab, setActiveTab] = useState<TabProps | string>(defaultActiveTab);
    const [teamOverviewData, setTeamOverviewData] = useState<TeamOverviewData>();
    const [teamOverviewTableData, setTeamOverviewTableData] = useState<TableData['data']>();
    const [teamOverviewTableTitle, setTeamOverviewTableTitle] = useState<string>(defaultActiveTab.title);
    const [error, setError] = useState<TeamOverviewError | undefined>();
    const [loading, setLoading] = useState<boolean>(true);

    useEffect(() => {
        getTeamOverview().then(response => {
            if ((response as TeamOverviewError).error) {
                setError(response as TeamOverviewError);
            }
            else {
                setTeamOverviewData(response as TeamOverviewData)
                setTeamOverviewTableData(prepTeamData(response as TeamOverviewData))
            }
        })
            .finally(() => setLoading(false))
            .catch(error => {
                setError(error.toString());
            });
    }, []);

    useEffect(() => {
        if (teamOverviewData) {
            setTeamOverviewTableData(prepTeamData(teamOverviewData))
        }
    }, [activeTab])

    const prepTeamData = (response: TeamOverviewData): TableData['data'] => {
        const team = (activeTab as TabProps)?.path ?? activeTab

        return response[team]._embedded.teams.map(team => ({
            id: team.uniform_name,
            'navn': renderTeamNameColumn(team),
            'teammedlemmer': team.team_user_count,
            'ansvarlig': team.manager.display_name.split(", ").reverse().join(" ")
        }));
    }

    const handleTabClick = (tab: string) => {
        setActiveTab(tab);
        if (tab === 'myTeams') {
            setTeamOverviewTableTitle('Mine team')
        } else {
            setTeamOverviewTableTitle('Alle teams')
        }
    };

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

    // TODO: Will be used by the other pages as well. Can be repurposed if necessary
    function renderErrorAlert() {
        return (
            <Dialog type='warning' title="Could not fetch teams">
                {error?.error.message}
            </Dialog >
        )
    }

    function renderSkeletonOnLoad() {
        return (
            <>
                <Skeleton variant="rectangular" animation="wave" height={60} />
                <Skeleton variant="text" animation="wave" sx={{ fontSize: '5.5rem' }} width={150} />
                <Skeleton variant="rectangular" animation="wave" height={200} />
            </>
        )
    }

    function renderContent() {
        if (error) {
            return renderErrorAlert()
        }

        if (loading) {
            return renderSkeletonOnLoad()
        }

        if (teamOverviewTableData) {
            const teamOverviewTableHeaderColumns = [{
                id: 'navn',
                label: 'Navn',
            },
            {
                id: 'teammedlemmer',
                label: 'Teammedlemmer',
            }, {
                id: 'ansvarlig',
                label: 'Ansvarlig'
            }]

            return (
                <>
                    <Tabs
                        onClick={handleTabClick}
                        activeOnInit={defaultActiveTab.path}
                        items={
                            [
                                { title: `Mine team (${teamOverviewData?.myTeams.count ?? 0})`, path: 'myTeams' },
                                { title: `Alle team (${teamOverviewData?.allTeams.count ?? 0})`, path: 'allTeams' },
                            ]}
                    />
                    <Divider dark />
                    <Title size={2} className={pageLayoutStyles.tableTitle}>{teamOverviewTableTitle}</Title>
                    <Table
                        columns={teamOverviewTableHeaderColumns}
                        data={teamOverviewTableData as TableData['data']}
                    />
                </>
            )
        }
    }

    return (
        <PageLayout
            title="Teamoversikt"
            content={renderContent()}
        />
    )
}