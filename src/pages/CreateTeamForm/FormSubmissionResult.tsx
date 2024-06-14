import styles from './createTeamForm.module.scss'
import { Dialog } from '@statisticsnorway/ssb-component-library'
import { Option as O } from 'effect'

export interface FormSubmissionResult {
    readonly success: boolean
    readonly message: string
}

export interface FormSubmissionResultProps {
    formSubmissionResult: O.Option<FormSubmissionResult>
}

const FormSubmissionResult = ({ formSubmissionResult }: FormSubmissionResultProps) =>
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

export default FormSubmissionResult
