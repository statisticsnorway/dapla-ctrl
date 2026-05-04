import styles from './teamDetail.module.scss'

import { TeamDetailData, Group } from '../../services/teamDetail'
import { User } from '../../services/teamMembers'
import SidebarModal, { SidebarHeader } from '../../components/SidebarModal/SidebarModal'

interface AddMember {
  loadingUsers: boolean
  setRefreshData: React.Dispatch<React.SetStateAction<boolean>>
  userData: User[] | undefined
  teamDetailData: TeamDetailData | undefined
  teamModalHeader: SidebarHeader
  teamGroups: Group[]
  open: boolean
  onClose: CallableFunction
}
const AddTeamMember = ({ teamDetailData, teamModalHeader, open, onClose }: AddMember) => {
  if (teamDetailData) {
    return (
      <SidebarModal
        open={open}
        onClose={() => onClose()}
        header={teamModalHeader}
        body={{
          modalBodyTitle: 'Legg person til teamet',
          modalBody: (
            <div className={styles.inputSpacing}>
              Frem til lansering av nye Dapla Ctrl kan du ikke legge til personer i teamet. Vennligst kom tilbake om 15
              minutter eller se Viva Engage under "Driftsmeldinger IT" for siste nytt.
            </div>
          ),
        }}
        footer={{}}
      />
    )
  }
  return
}

export default AddTeamMember
