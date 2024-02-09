import Breadcrumb from '../Breadcrumb'
import Header from '../Header/Header'
import SidebarModal from '../SidebarModal/SidebarModal'
import styles from './pagelayout.module.scss'

import { Title, LeadParagraph } from '@statisticsnorway/ssb-component-library'

interface PageLayoutProps {
  title: string
  description?: JSX.Element
  button?: JSX.Element
  content?: JSX.Element
}

export default function PageLayout({ title, description, button, content }: PageLayoutProps) {
  return (
    <>
      <Header />
      <main className={styles.container}>
        <Breadcrumb />
        {/* TODO: Remove after testing; or implement a temporary button that toggles this in one of the pages */}
        <SidebarModal
        header={{modalType: "Medlem", modalTitle: "Arbmark register", modalDescription: "arbmark-register"}}
        footer={{submitButtonText: "Legg til medlem", handleSubmit: () => {console.log("clicked")}}}
        />
        <div className={styles.title}>
          <Title size={1}>{title}</Title>
          {button}
        </div>
        <LeadParagraph>{description}</LeadParagraph>
        {content}
      </main>
    </>
  )
}
