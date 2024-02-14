import { User } from "../@types/user"

const DAPLA_TEAM_API_URL = import.meta.env.VITE_DAPLA_TEAM_API_URL
const USERS_URL=`${DAPLA_TEAM_API_URL}/users`

interface A { // find a suitable name
    user: string
    managedUsers: User[]

}

/*

manger: obr@ssb.no
users
    [
        {
            ...
            section_manager,
            teams,
            groups,
        },
        ...
    ]

*/


const fetchUsersManagedBy = async (principalName: string, token: string): Promise<A> => {
    const managedUsersUrl = new URL(`${USERS_URL}/${principalName}/managed-users`)
    managedUsersUrl.searchParams.set('embed', 'teams,groups,section_manager')
    const selectedProperties = [
        'display_name',
        'principal_name',
        'groups.uniform_name',
        'teams.uniform_name',
        'section_manager.display_name'
    ]

    selectedProperties.forEach((item) => managedUsersUrl.searchParams.append('select', item))
    return fetch(managedUsersUrl.toString(), {
        method: "GET",
        headers: {
            accept: '*/*',
            Authorization: `Bearer ${token}`,
            },
    }).then((response => {
        //TODO: aggregate response
        return response.json() as Promise<A>
    }))
}