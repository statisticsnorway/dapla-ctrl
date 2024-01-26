import PageLayout from '../components/PageLayout/PageLayout'
import Table from '../components/Table/Table'

import { useEffect, useState } from "react"
import { getAllTeams, getMyTeams, Root, TeamOverviewError, Team } from "../api/teamOverview"
import { Title, Dialog, Link } from "@statisticsnorway/ssb-component-library"

export default function TeamOverview() {
    const [allTeams, setAllTeams] = useState<Root | undefined>();
    const [error, setError] = useState<TeamOverviewError | undefined>();

    useEffect(() => {
        getAllTeams().then(response => {
            if ((response as TeamOverviewError).error) {
                setError(response as TeamOverviewError)
            } else {
                setAllTeams(response as Root);
            }
        }).catch(error => {
            setError(error.toString());
        });
    }, []);

    function renderTeamNameColumn(team: Team) {
        return (
            <>
                <span>
                    <Link href={""}>
                        <b>{team.uniformName}</b>
                    </Link>
                </span>
                {/* TODO: Fetch department from API. Teams are missing a department property */}
            </>
        )
    }

    function renderAllTeams() {
        if (error) {
            return (
                <Dialog type='warning' title="Could not fetch teams">
                    {error.error.message}
                </Dialog>
            )
        }

        if (allTeams && allTeams.count) {
            const allTeamsTableHeaderColumns = [{
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

            const allTeamsTableDataColumns = allTeams._embedded.teams.map(team => ({
                id: team.uniformName,
                'navn': renderTeamNameColumn(team),
                'teammedlemmer': team.teamUserCount,
                'ansvarlig': team.manager.displayName
            }))

            // TODO: Loading can be replaced by a spinner eventually
            return (
                <>
                    <Title size={2}>Alle team</Title>
                    <Table
                        columns={allTeamsTableHeaderColumns}
                        data={allTeamsTableDataColumns}
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
            content={renderAllTeams()}
        />
    )
}