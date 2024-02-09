import styles from './sidebar.module.scss'

import { useEffect, useRef, useState } from 'react'
import { Link, Button, Dialog, Input, Dropdown } from '@statisticsnorway/ssb-component-library'
import { X } from 'react-feather'

interface SidebarHeader {
  modalType: string
  modalTitle: string
  modalDescription: string
}

const SidebarModalHeader = ({ modalType, modalTitle, modalDescription }: SidebarHeader): JSX.Element => {
  return (
    <div className={styles.modalHeader}>
      <div className={styles.modalType}>
        <span>{modalType}</span>
      </div>
      <div className={styles.modalTitle}>
        <h1>{modalTitle}</h1>
      </div>
      <div className={styles.modalDescription}>
        <p>{modalDescription}</p>
      </div>
    </div>
  )
}

interface SidebarFooter {
  submitButtonText: string
  onClose?: () => void
  handleSubmit?: () => void
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

interface SidebarModal {
  open: boolean
  onClose: () => void
  header: SidebarHeader
  body: JSX.Element
  footer: SidebarFooter
}

const SidebarModal = ({ open, onClose, header, footer, body }: SidebarModal) => {
  return (
    <div className={`${styles.container} ${open ? styles.open : ''}`}>
      <div className={styles.header}>
        <button onClick={onClose}>
          <X className={styles.xIcon} size={32} />
        </button>
        {/* TODO: Should this be wrapped around button? If yes, remove default styling on button */}
      </div>
      <SidebarModalHeader {...header} />
      <div className={styles.body}>
        {/* Form goes here */}
        {body}
      </div>
      <SidebarModalFooter {...footer} onClose={footer.onClose ? footer.onClose : onClose} />
    </div>
  )
}

export default SidebarModal
