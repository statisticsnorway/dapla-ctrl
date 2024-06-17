import styles from './createTeamForm.module.scss'
import { Dialog } from '@statisticsnorway/ssb-component-library'
import { Option as O } from 'effect'
import { CircularProgress } from '@mui/material'

export interface FormSubmissionResult {
  readonly success: boolean
  readonly message: string
}

export interface FormSubmissionResultProps {
  loading: boolean
  formSubmissionResult: O.Option<FormSubmissionResult>
}

// <Skeleton variant='rectangular' animation='wave' width={380} height={120} />
const FormSubmissionResult = ({ formSubmissionResult, loading }: FormSubmissionResultProps) =>
  loading ? (
    <div className={styles.center}>
      <CircularProgress />
    </div>
  ) : (
    O.match(formSubmissionResult, {
      onNone: () => undefined,
      onSome: (res) => {
        const title = res.success ? 'Skjema ble innsendt' : 'Feil oppstod ved innsending av skjema'
        const dialogType = res.success ? 'info' : 'warning'
        return (
          <Dialog className={styles.warning} type={dialogType} title={title}>
            {res.message}
          </Dialog>
        )
      },
    })
  )

export default FormSubmissionResult
