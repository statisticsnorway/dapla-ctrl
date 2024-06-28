import { Dialog, Link } from '@statisticsnorway/ssb-component-library'
import { useLocation } from 'react-router-dom'

import PageLayout from '../../components/PageLayout/PageLayout'
import styles from './createTeamForm.module.scss'

export interface TeamCreatedProps {
  kubenPullRequestUrl: string
}

const TeamCreated = () => {
  const { state }: { state: TeamCreatedProps } = useLocation()
  const renderContent = () => (
    <Dialog className={styles.warning} type={'info'} title={'Skjema er sendt inn'}>
      <span>{`Team Skyinfrastruktur vil gi beskjed til seksjonsleder når teamet er opprettet og klar til bruk. Søknaden er `}</span>
      <Link href={state.kubenPullRequestUrl}>dokumentert på GitHub. </Link>
      <span>{'Opprett en Kundeservice-sak hvis du ønsker å gjøre endringer før teamet er opprettet.'}</span>
    </Dialog>
  )

  return <PageLayout title='Opprett Team' content={renderContent()} />
}

export default TeamCreated
