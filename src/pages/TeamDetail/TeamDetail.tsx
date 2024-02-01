
import { useEffect, useState } from 'react'
import PageLayout from '../../components/PageLayout/PageLayout'
import { TeamDetailData, getTeamDetail } from '../../api/teamDetail'
import { useParams } from 'react-router-dom'
import { ErrorResponse } from '../../@types/error'


export default function TeamOverview() {
    const { teamId } = useParams<{ teamId: string }>()
    const [teamDetailData, setTeamDetailData] = useState<TeamDetailData>()
    const [error, setError] = useState<ErrorResponse | undefined>()
    const [loading, setLoading] = useState<boolean>(true)

    useEffect(() => {
        if (!teamId) return
        console.log(teamId)
        getTeamDetail(teamId)
            .then((response) => {
                if ((response as ErrorResponse).error) {
                    console.log(response)
                    setError(response as ErrorResponse)
                } else {
                    console.log((response))
                    setTeamDetailData(response as TeamDetailData)
                }
            })
            .finally(() => setLoading(false))
            .catch((error) => {
                setError(error.toString())
            })

    }, [teamId])


    return <PageLayout title='Teamoversikt' content={<h1>TeamDetail</h1>} />
}
