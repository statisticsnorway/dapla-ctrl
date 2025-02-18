import { JobResponse } from '../../services/teamDetail'
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

export const renderSidebarModalWarning = (errorList: JobResponse[]) => {
  if (!errorList.length) return null

  return (
    <Dialog type="warning" title="API-feil oppstod">
      <div>
        <p>Det oppstod følgende feil under forespørselen:</p>
        <ul>
          {errorList.map((error, index) => (
            <li
              key={index}
            >
              <p>
                <strong>Error Code:</strong> {error.statusCode}
              </p>
              <p>
                <strong>Detaljer:</strong> {error.detail}
              </p>
            </li>
          ))}
        </ul>
      </div>
    </Dialog>
  )
}
