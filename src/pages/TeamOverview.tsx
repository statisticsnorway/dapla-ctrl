import PageLayout from '../components/PageLayout/PageLayout'
import Table, { TableData } from '../components/Table/Table'
import { useEffect, useState } from "react"
import { getTeams, Root, TeamOverviewError, Team, Path } from "../api/teamOverview"
import { Dialog, Link, Tabs, Divider } from "@statisticsnorway/ssb-component-library"

interface TabProps {
    title: string,
    path: string
}

export default function TeamOverview() {
    const defaultActiveTab = {
        title: 'Mine team',
        path: 'myTeams'
    }

    const [activeTab, setActiveTab] = useState<TabProps | string>(defaultActiveTab);
    const [teamOverviewTableData, setTeamOverviewTableData] = useState<TableData['data']>();
    const [error, setError] = useState<TeamOverviewError | undefined>();

    // initial page load
    useEffect(() => {
        getTeams('myTeams').then(response => {
            if ((response as TeamOverviewError).error) {
                setError(response as TeamOverviewError);
            }
            else {
                setTeamOverviewTableData(prepTeamData(response as Root));
            }
        }).catch(error => {
            setError(error.toString());
        });
    }, []);

    useEffect(() => {
        getTeams(((activeTab as TabProps)?.path ?? activeTab) as Path).then(response => {
            if ((response as TeamOverviewError).error) {
                setError(response as TeamOverviewError)
            } else {
                setTeamOverviewTableData(prepTeamData(response as Root));
            }
        }).catch(error => {
            setError(error.toString());
        });
    }, [activeTab])


    const prepTeamData = (response: Root): TableData['data'] => {
        return response._embedded.teams.map(team => ({
            id: team.uniformName,
            'navn': renderTeamNameColumn(team),
            'teammedlemmer': team.teamUserCount,
            'ansvarlig': team.manager.displayName
        }));
    }

    const handleTabClick = (tab: string) => {
        setActiveTab(tab);
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
                                // { title: `Mine team (${teamOverviewTableData ? myTeamsData.count : 0})`, path: 'myTeams' },
                                // { title: `Alle team (${teamOverviewTableData ? allTeamsData.count : 0})`, path: 'allTeams' },
                                // TODO: Add count
                                { title: `Mine team`, path: 'myTeams' },
                                { title: `Alle team`, path: 'allTeams' },
                            ]}
                    />
                    <Divider dark />
                    <Table
                        columns={teamOverviewTableHeaderColumns}
                        // TODO: Can be undefined:
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