import styles from './sidebar.module.scss'

import { useState } from 'react'
import { Link } from '@statisticsnorway/ssb-component-library'
import { X } from 'react-feather'

interface SidebarModal {
  header?: JSX.Element
  body?: JSX.Element
  button?: JSX.Element
}

const SidebarModal = ({ header, body, button }: SidebarModal) => {
  const [isOpen, setOpen] = useState<boolean>(true)

  const handleClose = () => {
    setOpen(!isOpen)
  }

  if (isOpen) {
    return (
      <div className={styles.container}>
        <div className={styles.header}>
          {/* TODO: Should this be wrapped around button? If yes, remove default styling on button */}
          <button onClick={handleClose}>
            <X className={styles.xIcon} size={32} />
          </button>
          {header}
        </div>
        <div className={styles.body}>
          {/* Form goes here */}
          {body}
        </div>
        <div className={styles.footer}>
          <Link onClick={handleClose}>Avbryt</Link>
          {button}
        </div>
      </div>
    )
  }
}

export default SidebarModal
