import styles from './modal.module.scss'

import { Modal as MUIModal } from '@mui/material'
import { Title, Divider } from '@statisticsnorway/ssb-component-library'

interface Modal {
  open: boolean
  onClose: CallableFunction
  modalTitle?: string
  body?: JSX.Element
  footer?: JSX.Element
}

const Modal = ({ open, onClose, modalTitle, body, footer }: Modal) => {
  return (
    <MUIModal open={open} onClose={() => onClose()}>
      <div className={styles.deleteConfirmationModalContainer}>
        {modalTitle && (
          <div className={styles.deleteConfirmationModalHeader}>
            <Title size={2}>{modalTitle}</Title>
          </div>
        )}
        {body && <div className={styles.deleteConfirmationModalBody}>{body}</div>}
        {footer && (
          <>
            <Divider light />
            <div className={styles.deleteConfirmationModalFooter}>{footer}</div>
          </>
        )}
      </div>
    </MUIModal>
  )
}

export default Modal
