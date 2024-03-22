import styles from './deletelink.module.scss'

import { Trash2 } from 'react-feather'

interface DeleteLink {
  children: string
  tabIndex?: number
  icon?: boolean
  handleDeleteUser: CallableFunction
}

const DeleteLink = ({ children, tabIndex, icon, handleDeleteUser }: DeleteLink) => {
  return (
    <a className={styles.deleteLinkWrapper} tabIndex={tabIndex ?? 0} onClick={() => handleDeleteUser}>
      {icon && <Trash2 size={22} />}
      <span>{children}</span>
    </a>
  )
}

export default DeleteLink
