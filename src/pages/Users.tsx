import { PageLayout } from "../components/PageLayout/PageLayout"
import { getAllTeams, ApiResponse } from "../api/teamApi"
import { useEffect, useState } from "react"
import styles from './Users.module.scss';

export function Users() {
    // get token from localstorage
    const [teams, setTeams] = useState<ApiResponse | undefined>();

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
            {teams && teams.data._embedded.teams.length > 0 ? (
                <div>
                    <h2>Team List</h2>
                    <table>
                        <thead>
                            <tr>
                                <th>Uniform Name</th>
                                <th>Display Name</th>
                            </tr>
                        </thead>
                        <tbody>
                            {teams.data._embedded.teams.map(team => (
                                <tr key={team.uniformName}>
                                    <td>{team.uniformName}</td>
                                    <td>{team.displayName}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            ) : (
                <div>No teams found.</div>
            )}

        </>
    )
}
