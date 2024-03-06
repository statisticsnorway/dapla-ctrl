import Breadcrumb from '../Breadcrumb'
import Header from '../Header/Header'
import styles from './pagelayout.module.scss'

import { Title } from '@statisticsnorway/ssb-component-library'

interface PageLayoutProps {
  title?: string | JSX.Element
  button?: JSX.Element
  content?: JSX.Element
}

const PageLayout = ({ title, button, content }: PageLayoutProps) => {
  return (
    <>
      <Header />
      <main className={styles.container}>
        <Breadcrumb />
        <div className={styles.titleContainer}>
          {title && (
            <Title size={1} className={styles.title}>
              {title}
            </Title>
          )}
          {button && <div className={styles.button}>{button}</div>}
        </div>
        {content}
      </main>
    </>
  )
}

export default PageLayout
