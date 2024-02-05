import Breadcrumb from '../Breadcrumb'
import Header from '../Header/Header'
import styles from './pagelayout.module.scss'

import { Title } from '@statisticsnorway/ssb-component-library'

interface PageLayoutProps {
  title?: string
  button?: JSX.Element
  content?: JSX.Element
}

export default function PageLayout({ title, button, content }: PageLayoutProps) {
  return (
    <>
      <Header />
      <main className={styles.container}>
        <Breadcrumb />
        <div className={styles.title}>
          {title && <Title size={1}>{title}</Title>}
          {button}
        </div>
        {content}
      </main>
    </>
  )
}
