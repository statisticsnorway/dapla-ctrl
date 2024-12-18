import { Team } from "../../src/services/teamDetail"
import { getTeamFromGroup } from "../../src/utils/utils"

describe('getTeamFromGroup', () => {
    it('should return the longest matching team name for a group name', () => {
      const allTeams = [
        { uniform_name: 'donald-du' },
        { uniform_name: 'donald-duck' },
        { uniform_name: 'mickey-mouse' },
      ]
      const groupName = 'donald-duck-data-admins'
  
      const result = getTeamFromGroup(allTeams, groupName)
  
      expect(result).toBe('donald-duck') // Longest match
    })
  
    it('should return an empty string when no matches are found', () => {
      const allTeams = [
        { uniform_name: 'mickey-mouse' },
        { uniform_name: 'goofy' },
      ]
      const groupName = 'donald-duck-data-admins'
  
      const result = getTeamFromGroup(allTeams, groupName)
  
      expect(result).toBe('') // No matches
    })
  
    it('should handle empty input arrays', () => {
      const allTeams: Team[] = []
      const groupName = 'donald-duck-data-admins'
  
      const result = getTeamFromGroup(allTeams, groupName)
  
      expect(result).toBe('') // Empty array
    })
  })
  