export interface User {
    principal_name: string
    azure_ad_id: string
    display_name: string
    first_name: string
    last_name: string
    email: string
    manager?: User
    photo?: string
}