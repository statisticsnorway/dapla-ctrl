import pageLayoutStyles from '../../components/PageLayout/pagelayout.module.scss'

import { useEffect, useState } from "react"
import PageLayout from '../../components/PageLayout/PageLayout'
import { TabProps } from '../../@types/pageTypes'
import Table, { TableData } from '../../components/Table/Table'
import { getTeamOverview, TeamOverviewData, TeamOverviewError, Team } from "../../api/teamOverview"
import { Dialog, Title, Link, Tabs, Divider } from "@statisticsnorway/ssb-component-library"

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

    // initial page load
    useEffect(() => {
        getTeamOverview().then(response => {
            if ((response as TeamOverviewError).error) {
                setError(response as TeamOverviewError);
            }
            else {
                setTeamOverviewData(response as TeamOverviewData)
                setTeamOverviewTableData(prepTeamData(response as TeamOverviewData))
            }
        }).catch(error => {
            setError(error.toString());
        });
    }, []);

    useEffect(() => {
        // TODO: Add loading
        if (teamOverviewData) {
            setTeamOverviewTableData(prepTeamData(teamOverviewData))
        }
    }, [activeTab])

    const prepTeamData = (response: TeamOverviewData): TableData['data'] => {
        const team = (activeTab as TabProps)?.path ?? activeTab

        return response[team]._embedded.teams.map(team => ({
            id: team.uniformName,
            'navn': renderTeamNameColumn(team),
            'teammedlemmer': team.teamUserCount,
            'ansvarlig': team.manager.displayName
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
                    <Link href={`/${team.uniformName}`}>
                        <b>{team.uniformName}</b>
                    </Link>
                </span>
            </>
        )
    }

    function renderContent() {
        if (error) {
            return (
                <Dialog type='warning' title="Could not fetch teams">
                    {error.error.message}
                </Dialog>
            )
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

            // TODO: Loading can be replaced by a spinner eventually
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
            ) || <p>Loading...</p> || (
                    <p>No teams found.</p>
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