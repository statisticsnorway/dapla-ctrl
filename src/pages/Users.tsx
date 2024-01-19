import PageLayout from "../components/PageLayout/PageLayout"
import { getAllTeams, TeamApiResponse } from "../api/teamApi"
import { useEffect, useState } from "react"
import { Dialog } from "@statisticsnorway/ssb-component-library"

export default function Users() {
    const [teams, setTeams] = useState<TeamApiResponse | undefined>();
    const [error, setError] = useState<string | undefined>();

    useEffect(() => {
        getAllTeams().then(response => {
            setTeams(response);
        }).catch(error => {
            setError(error.toString());
        });
    }, []);

    return (
        <>
            <PageLayout
                title="Medlemmer"
            />
            {error ? <Dialog type='warning' title="Could not fetch teams">
                {error}
            </Dialog> : teams && teams.data.length > 0 && (
                <>
                    <h2>Team List</h2>
                    <table>
                        <thead>
                            <tr>
                                <th>Uniform Name</th>
                                <th>Display Name</th>
                            </tr>
                        </thead>
                        <tbody>
                            {teams.data.map(team => (
                                <tr key={team.uniformName}>
                                    <td>{team.uniformName}</td>
                                    <td>{team.displayName}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </>
            ) || <p>Loading...</p> || (
                <p>No teams found.</p>
            )}
        </>
    )
}
