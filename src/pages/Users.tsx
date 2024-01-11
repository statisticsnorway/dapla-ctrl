import { PageLayout } from "../components/PageLayout/PageLayout"
import { RootObject } from "../temp/TeamResponseObject";
import React from 'react';

async function listTeams(): Promise<RootObject> {
    try {
        const response = await fetch(import.meta.env.VITE_DAPLA_TEAM_API_URL, {
            method: "GET",
            headers: {
                "accept": "*/*",
                "Authorization": "Bearer " + import.meta.env.VITE_DAPLA_TEAM_API_BEARER_TOKEN,
            }
        });
        if (!response.ok) {
            throw new Error("Response is not ok");
        }
        const data: RootObject = await response.json();
        return data;
    } catch (error) {
        console.error('Error:', error);
        throw error;
    }
}

export function Users(): JSX.Element {
    const [data, setData] = React.useState<RootObject | null>(null);
    const [loading, setLoading] = React.useState(true);
    const [error, setError] = React.useState<Error | null>(null);

    React.useEffect(() => {
        listTeams().then(teams => {
            setData(teams);
            setLoading(false);
        }).catch(err => {
            setError(err);
            setLoading(false);
        });
    }, []);

    if (loading) {
        return <div>Loading...</div>;
    }

    if (error) {
        return <div>Error: {error.message}</div>;
    }


    console.log(data)

    return (
        <>
            <PageLayout
                title="Members"
                buttonText="Add Member"
            />
            {data && data._embedded.teams.length > 0 ? (
                <div>
                    <h2>Team List</h2>
                    <ul>
                        {data._embedded.teams.map(team => (
                            <li key={team.uniformName}>{team.displayName}</li>
                        ))}
                    </ul>
                </div>
            ) : (
                <div>No teams found.</div>
            )}
        </>
    );
}