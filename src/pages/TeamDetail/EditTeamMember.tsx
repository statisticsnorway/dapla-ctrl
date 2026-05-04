import styles from './teamDetail.module.scss'

import { UserInfo } from './TeamDetail'
import { TeamDetailData, Group } from '../../services/teamDetail'
import SidebarModal, { SidebarHeader } from '../../components/SidebarModal/SidebarModal'

interface EditTeamMember {
  editUserInfo: UserInfo
  setRefreshData: React.Dispatch<React.SetStateAction<boolean>>
  teamDetailData: TeamDetailData | undefined
  teamModalHeader: SidebarHeader
  teamGroups: Group[]
  open: boolean
  onClose: CallableFunction
}

const EditTeamMember = ({ editUserInfo, teamDetailData, teamModalHeader, open, onClose }: EditTeamMember) => {
  if (teamDetailData && editUserInfo) {
    return (
      <SidebarModal
        open={open}
        onClose={() => {
          onClose()
        }}
        header={teamModalHeader}
        footer={{}}
        body={{
          modalBodyTitle: `Endre tilgang til "${editUserInfo.name}"`,
          modalBody: (
            <>
              <div className={styles.modalBodyDialog}>
                Frem til lansering av nye Dapla Ctrl kan du ikke legge til personer i teamet. Vennligst kom tilbake om
                15 minutter eller se Viva Engage under "Driftsmeldinger IT" for siste nytt.
              </div>
            </>
          ),
        }}
      />
    )
  }
}

export default EditTeamMember
