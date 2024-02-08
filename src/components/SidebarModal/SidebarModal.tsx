import styles from './sidebar.module.scss'

import { useState } from 'react'
import { Title, Link, Button } from '@statisticsnorway/ssb-component-library'
import { X } from 'react-feather'

interface SidebarHeader {
  modalType: string
  modalTitle: string
  modalDescription: string
}

const SidebarModalHeader = ({modalType, modalTitle, modalDescription}: SidebarHeader): JSX.Element => {
  return (
    <div className={styles.modalHeader}>
      <div className={styles.modalType}><span>{modalType}</span></div>
      <div className={styles.modalTitle}><h1>{modalTitle}</h1></div>
      <div className={styles.modalDescription}><p>{modalDescription}</p></div>
    </div>
  )
}

interface SidebarFooter {
  submitButtonText: string
  handleClose?: () => void
}

const SidebarModalFooter = ({submitButtonText, handleClose}: SidebarFooter): JSX.Element => {
  return (
    <div className={styles.modalFooter}>
      <Link onClick={handleClose}>Avbryt</Link>
      <div className={styles.modalFooterButtonText}><Button primary>{submitButtonText}</Button></div>
    </div>
  )
}

interface SidebarModal {
  header: SidebarHeader
  body?: JSX.Element
  footer: SidebarFooter
}

const SidebarModal = ({ header, footer, body }: SidebarModal) => {
  const [isOpen, setOpen] = useState<boolean>(true)

  const handleClose = () => {
    setOpen(!isOpen)
  }

  if (isOpen) {
    return (
      <div className={styles.container}>
        <div className={styles.header}>
          <button onClick={handleClose}>
            <X className={styles.xIcon} size={32} />
          </button>
          {/* TODO: Should this be wrapped around button? If yes, remove default styling on button */}
          <SidebarModalHeader {...header} />
        </div>
        <div className={styles.body}>
          {/* Form goes here */}
          {body}
        </div>
        <SidebarModalFooter {...footer} handleClose={footer.handleClose ? footer.handleClose : handleClose} />
      </div>
    )
  }
}

export default SidebarModal
