export interface User {
    principal_name: string
    azure_ad_id: string
    display_name: string
    first_name: string
    last_name: string
    email: string
    division_name?: string
    division_code?: number
    section_name?: string
    section_code?: number
    manager?: User
    photo?: string
}