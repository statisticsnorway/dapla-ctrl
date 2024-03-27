import styles from './sidebarmodal.module.scss'

import { Title, Link, Button } from '@statisticsnorway/ssb-component-library'
import { X } from 'react-feather'

import { Drawer } from '@mui/material'

export interface SidebarHeader {
  modalType?: string
  modalTitle: string
  modalDescription?: string
}

interface SidebarBody {
  modalBodyTitle: string
  modalBody: JSX.Element
}

interface SidebarFooter {
  submitButtonText: string
  onClose?: () => void
  handleSubmit?: () => void
}

interface SidebarModal {
  open: boolean
  onClose: () => void
  header: SidebarHeader
  body: SidebarBody
  footer: SidebarFooter
}

const SidebarModalHeader = ({ modalType, modalTitle, modalDescription }: SidebarHeader): JSX.Element => {
  return (
    <div className={styles.modalHeader}>
      {modalType && <span>{modalType}</span>}
      {<Title size={1}>{modalTitle}</Title>}
      {modalDescription && <p>{modalDescription}</p>}
    </div>
  )
}

const SidebarModalBody = ({ modalBodyTitle, modalBody }: SidebarBody): JSX.Element => {
  return (
    <div className={styles.modalBody}>
      <Title size={2}>{modalBodyTitle}</Title>
      {modalBody}
    </div>
  )
}

const SidebarModalFooter = ({ submitButtonText, onClose, handleSubmit }: SidebarFooter): JSX.Element => {
  return (
    <div className={styles.modalFooter}>
      <Link onClick={onClose}>Avbryt</Link>
      <div className={styles.modalFooterButtonText}>
        <Button onClick={handleSubmit} primary>
          {submitButtonText}
        </Button>
      </div>
    </div>
  )
}

const SidebarModal = ({ open, onClose, header, footer, body }: SidebarModal) => {
  return (
    <Drawer
      anchor='right'
      transitionDuration={300} // milliseconds
      elevation={0} // no shadow styling
      ModalProps={{ BackdropProps: { invisible: true } }}
      open={open}
      onClose={onClose}
    >
      <div className={styles.modalContainer}>
        <div>
          <button className={styles.closeButton} onClick={onClose}>
            <X className={styles.xIcon} size={32} />
          </button>
        </div>
        <SidebarModalHeader {...header} />
        <SidebarModalBody {...body} />
        <SidebarModalFooter {...footer} onClose={footer.onClose ? footer.onClose : onClose} />
      </div>
    </Drawer>
  )
}

export default SidebarModal
