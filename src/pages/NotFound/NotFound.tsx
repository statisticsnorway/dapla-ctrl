import notFoundStyles from './notFound.module.scss'

import { Title } from '@statisticsnorway/ssb-component-library'

export default function NotFound() {
  return (
    <>
      <Title size={1} className={notFoundStyles.Title}>404</Title>
      <Title size={2} className={notFoundStyles.SubTitle}>Not Found</Title>
    </>
  )
}
