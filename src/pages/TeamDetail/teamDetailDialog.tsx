import styles from './teamDetail.module.scss'

import { Dialog } from '@statisticsnorway/ssb-component-library'

export const renderSidebarModalInfo = (children: JSX.Element) => {
  return (
    <div className={styles.modalBodyDialog}>
      <Dialog type='info'>Det tar 1-2 minutter før tilgangen er aktivert og tabellen er oppdatert.</Dialog>
      {children}
    </div>
  )
}

export const renderSidebarModalWarning = (errorList: string[]) => {
  if (errorList.length) {
    return (
      <Dialog type='warning'>
        {typeof errorList === 'string' ? (
          errorList
        ) : (
          <ul>
            {errorList.map((errors) => (
              <li>{errors}</li>
            ))}
          </ul>
        )}
      </Dialog>
    )
  }
}
