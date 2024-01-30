import { Dialog } from '@statisticsnorway/ssb-component-library';
import PageLayout from "../components/PageLayout/PageLayout"
import { useContext, useEffect, useState } from "react"
import { DaplaCtrlContext } from "../provider/DaplaCtrlProvider";

import { getUserProfile } from '../api/userProfile';

import { User } from "../@types/user";
import { useParams } from "react-router-dom";
import { ErrorResponse } from "../@types/error";
import { Skeleton } from '@mui/material';

export default function UserProfile() {
    const { setData } = useContext(DaplaCtrlContext);
    const [error, setError] = useState<ErrorResponse | undefined>();
    const [loading, setLoading] = useState<boolean>();
    const [userProfileData, setUserProfileData] = useState<User>();
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

    // NOTE: This is where we set the data to the context, so that the breadcrumb can access it
    useEffect(() => {
        if (userProfileData) {
            const displayName = userProfileData.display_name.split(', ').reverse().join(' ');
            userProfileData.display_name = displayName;
            setData({ "displayName": displayName });
        }
    }, [userProfileData]);

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

        return (
            <>
                <h1>{userProfileData?.display_name}</h1>
            </>
        )
    }

    return (
        <PageLayout
            title="Teammedlemmer"
            content={renderContent()}
        />
    )
}
