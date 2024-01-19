import Cookies from 'js-cookie'
import PageLayout from '../components/PageLayout/PageLayout'
import Table from '../components/Table/Table'

import { useEffect, useState } from "react"
import { getAllTeams, TeamApiResponse, Team } from "../api/teamApi"
import { Title, Dialog, Link } from "@statisticsnorway/ssb-component-library"

export default function Home() {
    const [teams, setTeams] = useState<TeamApiResponse | undefined>();
    const [error, setError] = useState<string | undefined>();

    useEffect(() => {
        const token = Cookies.get('token');
        
        getAllTeams(token).then(response => {
            setTeams(response);
        }).catch(error => {
            setError(error.toString());
        });
    }, []);

    function renderTeamNameColumn(team: Team) {
        return (
            <>
                <span>
                    <Link href={team._links.self.href}>
                        <b>{team.displayName}</b>
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
                    {error}
                </Dialog>
            )
        }

        if (teams && teams.data.length) {
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
            
            const allTeamsTableDataColumns = teams.data.map(team => ({
                id: team.uniformName,
                'navn': renderTeamNameColumn(team),
                // TODO: Fetch team user count from API e.g. /teams/{team.uniformName}/users and users.count
                'teammedlemmer': 12,
                // TODO: 
                // * Fetch team manager? from API e.g. /groups/{team.uniformName}-managers/users and use users.displayName
                'ansvarlig': 'Lorem ipsum', 
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