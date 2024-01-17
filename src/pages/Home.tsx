import PageLayout from '../components/PageLayout/PageLayout'
import Table from '../components/Table/Table'

import { useEffect, useState } from "react"
import { getAllTeams, TeamApiResponse, Team } from "../api/teamApi"
import { Title, Dialog, Link } from "@statisticsnorway/ssb-component-library"

export default function Home() {
    const [teams, setTeams] = useState<TeamApiResponse | undefined>();
    const [error, setError] = useState<string | undefined>();

    useEffect(() => {
        const token = localStorage.getItem('token') as string;
        getAllTeams(token).then(response => {
            setTeams(response);
        }).catch(error => {
            setError(error.toString());
        });
    }, []);

    function renderTeamNameCell(team: Team) {
        return (
            <Link href={team._links.self.href}>{team.displayName}</Link>
        )
    }

    function renderAllTeams() {
        if(error) {
            return (
                <Dialog type='warning' title="Could not fetch teams">
                    {error}
                </Dialog>
            )
        }

        if (teams && teams.data.length) {
            const allTeamsTableHeader = [{
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
            
            const allTeamsTableData = teams.data.map(team => ({
                id: team.uniformName,
                'navn': renderTeamNameCell(team)
            }))
            
            return (
                <>
                    <Title size={2}>Alle team</Title>
                    <Table
                        columns={allTeamsTableHeader}
                        data={allTeamsTableData}
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