import { PageLayout } from "../components/PageLayout/PageLayout"
import { getAllTeams, TeamApiResponse } from "../api/teamApi"
import { useEffect, useState } from "react"

export function Users() {
    const [teams, setTeams] = useState<TeamApiResponse | undefined>();

    useEffect(() => {
        const token = localStorage.getItem('token') as string;
        getAllTeams(token).then(response => {
            setTeams(response);
        });
    }, []);

    return (
        <>
            <PageLayout
                title="Medlemmer"
                buttonText="Legg til medlem"
            />
            {teams && teams.data.length > 0 && (
                <div className="container">
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
                </div>
            ) || <div className="container">Loading...</div> || (
                    <div className="container">No teams found.</div>
                )}

        </>
    )
}
