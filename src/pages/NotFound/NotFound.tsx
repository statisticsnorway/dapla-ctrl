import notFoundStyles from './notFound.module.scss'

import { Title } from '@statisticsnorway/ssb-component-library'

const NotFound = () => {
  return (
    <div className={notFoundStyles.centerDiv}>
      <Title size={1} className={notFoundStyles.title}>
        Page not found!
      </Title>
      <Title size={2} className={notFoundStyles.subTitle}>
        Error code: 404
      </Title>
    </div>
  )
}

export default NotFound
