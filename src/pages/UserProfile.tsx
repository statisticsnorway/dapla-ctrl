import pageLayoutStyles from '../components/PageLayout/pagelayout.module.scss'

import { Dialog, Tabs, Divider, Title, Text, Link } from '@statisticsnorway/ssb-component-library';
import Table, { TableData } from '../components/Table/Table';
import PageLayout from "../components/PageLayout/PageLayout"
import { useContext, useEffect, useState } from "react"
import { DaplaCtrlContext } from "../provider/DaplaCtrlProvider";

import { getUserProfile, getUserTeamsWithGroups } from '../api/userProfile';

import { User } from "../@types/user";
import { Team } from "../@types/team";
import { useParams } from "react-router-dom";
import { ErrorResponse } from "../@types/error";
import { Skeleton } from '@mui/material';

export default function UserProfile() {
    const { setData } = useContext(DaplaCtrlContext);
    const [error, setError] = useState<ErrorResponse | undefined>();
    const [loading, setLoading] = useState<boolean>();
    const [userProfileData, setUserProfileData] = useState<User>();
    const [userTeams, setUserTeams] = useState<Team[]>();
    const [teamUserProfileTableData, setUserProfileTableData] = useState<TableData['data']>();
    const { principalName } = useParams();

    useEffect(() => {
        getUserProfile(principalName as string).then(response => {
            if ((response as ErrorResponse).error) {
                setError(response as ErrorResponse);
            } else {
                setUserProfileData(response as User);
            }
        }).finally(() => setLoading(false))
            .catch((error) => {
                setError({ error: { message: error.message, code: "500" } });
            })
    }, []);

    useEffect(() => {
        getUserTeamsWithGroups(principalName as string).then(response => {
            if ((response as ErrorResponse).error) {
                setError(response as ErrorResponse);
            } else {
                console.log(response);
                setUserTeams(response as Team[]);
                setUserProfileTableData(prepTeamData(response as Team[]));
            }
        }).finally(() => setLoading(false))
            .catch((error) => {
                setError({ error: { message: error.message, code: "500" } });
            })
    }, [])

    // NOTE: This is where we set the data to the context, so that the breadcrumb can access it
    useEffect(() => {
        if (userProfileData) {
            const displayName = userProfileData.display_name.split(', ').reverse().join(' ');
            userProfileData.display_name = displayName;
            setData({ "displayName": displayName });
        }
    }, [userProfileData]);

    const prepTeamData = (response: Team[]): TableData['data'] => {
        return response.teams.map(team => ({
            id: team.uniform_name,
            'navn': renderTeamNameColumn(team),
            'gruppe': team.groups.map(group => group.display_name).join(", "),
            'epost': userProfileData?.principal_name,
            'ansvarlig': team.manager.display_name.split(", ").reverse().join(" ")
        }));
    }

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

    function renderErrorAlert() {
        return (
            <Dialog type='warning' title="Could not fetch data">
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
        if (error) return renderErrorAlert();
        if (loading) return renderSkeletonOnLoad();

        if (teamUserProfileTableData) {
            const teamOverviewTableHeaderColumns = [{
                id: 'navn',
                label: 'Navn',
            },
            {
                id: 'gruppe',
                label: 'Gruppe',
            }, {
                id: 'epost',
                label: 'Epost'
            },
            {
                id: 'ansvarlig',
                label: 'Ansvarlig'
            }];
            return (
                <>
                    <h1>{userProfileData?.section_name}</h1>
                    <h1>{userProfileData?.display_name}</h1>

                    <Divider dark />
                    <Title size={2} className={pageLayoutStyles.tableTitle}>Hei</Title>
                    <Table
                        columns={teamOverviewTableHeaderColumns}
                        data={teamUserProfileTableData as TableData['data']}
                    />


                </>
            )
        }
    }

    return (
        <PageLayout
            title="Teammedlemmer"
            content={renderContent()}
        />
    )
}
