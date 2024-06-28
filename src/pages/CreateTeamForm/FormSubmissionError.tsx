import styles from './createTeamForm.module.scss'
import { Dialog } from '@statisticsnorway/ssb-component-library'
import { Option as O } from 'effect'
import { CircularProgress } from '@mui/material'

export interface FormSubmissionErrorProps {
  formSubmissionError: O.Option<string>
}

const FormSubmissionError = ({ formSubmissionError }: FormSubmissionErrorProps) =>
  O.match(formSubmissionError, {
    onNone: () => (
      <div className={styles.center}>
        <CircularProgress />
      </div>
    ),
    onSome: (errorMessage) => (
      <Dialog className={styles.warning} type='warning' title='Feil oppstod ved innsending av skjema'>
        {errorMessage}
      </Dialog>
    ),
  })

export default FormSubmissionError
