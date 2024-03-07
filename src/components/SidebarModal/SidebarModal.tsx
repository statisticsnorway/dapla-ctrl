import styles from './sidebar.module.scss'

import { useRef, useEffect } from 'react'
import { Title, Link, Button } from '@statisticsnorway/ssb-component-library'
import { X } from 'react-feather'

interface SidebarHeader {
  modalType?: string
  modalTitle: string
  modalDescription?: string
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
  body: JSX.Element
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
  // const [showScrollIndicator, setShowScrollIndicator] = useState(false)
  // const contentRef = useRef(null)

  // const checkForOverflow = () => {
  //   const element = contentRef.current
  //   if (!element) return

  //   // Check if the content is overflowing in the vertical direction
  //   const hasOverflow = element.scrollHeight > element.clientHeight
  //   setShowScrollIndicator(hasOverflow)
  // }

  // useEffect(() => {
  //   if (!open) return

  //   checkForOverflow()
  // }, [open])
  const sidebarModalRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleBackdropOnClick = (e: Event) => {
      if (sidebarModalRef.current && !sidebarModalRef?.current?.contains(e.target as Node)) onClose()
    }

    window.addEventListener('click', handleBackdropOnClick)
    return () => {
      window.removeEventListener('click', handleBackdropOnClick)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div className={`${styles.container} ${open ? styles.open : ''}`} ref={sidebarModalRef}>
      <div>
        <button className={styles.closeButton} onClick={onClose}>
          <X className={styles.xIcon} size={32} />
        </button>
      </div>
      <SidebarModalHeader {...header} />
      {/*<div className={styles.body} ref={contentRef} onScroll={checkForOverflow}> */}
      <div className={styles.modalBody}>
        {/* showScrollIndicator && <div className={styles.scroll}>↓ Scroll for å vise mer innhold</div> */}
        {body}
      </div>
      <SidebarModalFooter {...footer} onClose={footer.onClose ? footer.onClose : onClose} />
    </div>
  )
}

export default SidebarModal
