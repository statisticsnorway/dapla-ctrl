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
    <Dialog className={styles.warning} type={'info'} title={'Skjema ble innsendt'}>
      <span>{`Opprettelse av team ble registert. Prosessen kan følges `}</span>
      <Link href={state.kubenPullRequestUrl}>her</Link>
    </Dialog>
  )

  return <PageLayout title='Opprett Team' content={renderContent()} />
}

export default TeamCreated
