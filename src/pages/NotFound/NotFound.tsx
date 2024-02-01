import notFoundStyles from './notFound.module.scss'

import { Title } from '@statisticsnorway/ssb-component-library'

export default function NotFound() {
  return (
    <div className={notFoundStyles.CenterDiv}>
      <Title size={1} className={notFoundStyles.Title}>
        Page not found!
      </Title>
      <Title size={2} className={notFoundStyles.SubTitle}>
        Error code: 404
      </Title>
    </div>
  )
}
