import pageLayoutStyles from '../components/PageLayout/pagelayout.module.scss'

import { Dialog, Title, Text, Link } from '@statisticsnorway/ssb-component-library';
import Table, { TableData } from '../components/Table/Table';
import PageLayout from "../components/PageLayout/PageLayout"
import { useContext, useEffect, useState } from "react"
import { DaplaCtrlContext } from "../provider/DaplaCtrlProvider";
import { getGroupType } from "../utils/utils";

import { getUserProfile, getUserTeamsWithGroups, UserProfileTeamResult } from '../api/userProfile';

import { User } from "../@types/user";
import { Team } from "../@types/team";
import { useLocation, useParams } from "react-router-dom";
import { ErrorResponse } from "../@types/error";
import { Skeleton } from '@mui/material';

export default function UserProfile() {
    const { setData } = useContext(DaplaCtrlContext);
    const [error, setError] = useState<ErrorResponse | undefined>();
    const [loadingUserProfileData, setLoadingUserProfileData] = useState<boolean>(true);
    const [loadingTeamData, setLoadingTeamDataInfo] = useState<boolean>(true);
    const [userProfileData, setUserProfileData] = useState<User>();
    const [teamUserProfileTableData, setUserProfileTableData] = useState<TableData['data']>();
    const { principalName } = useParams();
    const location = useLocation();

    useEffect(() => {
        getUserProfile(principalName as string).then(response => {
            if ((response as ErrorResponse).error) {
                setError(response as ErrorResponse);
            } else {
                setUserProfileData(response as User);
            }
        }).finally(() => setLoadingUserProfileData(false))
            .catch((error) => {
                setError({ error: { message: error.message, code: "500" } });
            })
    }, [location, principalName]);

    useEffect(() => {
        if (userProfileData) {
            getUserTeamsWithGroups(principalName as string).then(response => {
                if ((response as ErrorResponse).error) {
                    setError(response as ErrorResponse);
                } else {
                    setUserProfileTableData(prepTeamData(response as UserProfileTeamResult));
                }
            }).finally(() => setLoadingTeamDataInfo(false))
                .catch((error) => {
                    setError({ error: { message: error.message, code: "500" } });
                })
        }
    }, [userProfileData])

    // required for breadcrumb
    useEffect(() => {
        if (userProfileData) {
            const displayName = userProfileData.display_name.split(', ').reverse().join(' ');
            userProfileData.display_name = displayName;
            setData({ "displayName": displayName });
        }
    }, [userProfileData]);

    const prepTeamData = (response: UserProfileTeamResult): TableData['data'] => {
        return response.teams.map(team => ({
            id: team.uniform_name,
            'navn': renderTeamNameColumn(team),
            'gruppe': team.groups?.map(group => getGroupType(group)).join(', '),
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
                <Skeleton variant="text" animation="wave" sx={{ fontSize: '5.5rem' }} width={150} />
                <Skeleton variant="rectangular" animation="wave" height={200} />
            </>
        )
    }

    function renderContent() {
        if (error) return renderErrorAlert();
        // TODO: cheesy method to exclude showing skeleton for profile information (username etc..)
        if (loadingTeamData && !loadingUserProfileData) return renderSkeletonOnLoad();

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
                label: 'Epost ?'
            },
            {
                id: 'ansvarlig',
                label: 'Ansvarlig'
            }];
            return (
                <>
                    <Title size={2} className={pageLayoutStyles.tableTitle}>Team</Title>
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
            title={userProfileData?.display_name as string}
            content={renderContent()}
            description={
                <>
                    <p>{userProfileData?.section_name}</p>
                    <p>{userProfileData?.principal_name}</p>
                </>
            }
        />
    )
}
