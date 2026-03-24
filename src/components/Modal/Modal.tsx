import styles from './modal.module.scss'

import { Modal as MUIModal } from '@mui/material'
import { Title, Divider } from '@statisticsnorway/ssb-component-library'
import type { ReactNode } from 'react'

interface Modal {
  open: boolean
  onClose: CallableFunction
  modalTitle?: string | ReactNode
  body?: ReactNode
  footer?: ReactNode
}

const Modal = ({ open, onClose, modalTitle, body, footer }: Modal) => {
  return (
    <MUIModal open={open} onClose={() => onClose()}>
      <div className={styles.modalContainer}>
        {modalTitle && (
          <div className={styles.modalHeader}>
            <Title size={2}>{modalTitle}</Title>
          </div>
        )}
        {body && <div className={styles.modalBody}>{body}</div>}
        {footer && (
          <>
            <Divider light />
            <div className={styles.modalFooter}>{footer}</div>
          </>
        )}
      </div>
    </MUIModal>
  )
}

export default Modal
