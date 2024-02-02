import { User } from '../@types/user'

export interface Group {
  uniform_name: string
  display_name: string
  manager?: User
}
