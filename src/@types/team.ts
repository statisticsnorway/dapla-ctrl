import { User } from './user'

export interface Team {
    uniform_name: string
    display_name: string
    division_name: string
    section_name: string
    section_code: number
    team_user_count: number
    manager: User
    groups?: string[]
}