import Breadcrumb from '../Breadcrumb'
import Header from '../Header/Header'
import SidebarModal from '../SidebarModal/SidebarModal'
import styles from './pagelayout.module.scss'

import { useState } from 'react'

import { Title, LeadParagraph } from '@statisticsnorway/ssb-component-library'

interface PageLayoutProps {
  title: string
  description?: JSX.Element
  button?: JSX.Element
  content?: JSX.Element
}

export default function PageLayout({ title, description, button, content }: PageLayoutProps) {
  const [isSidebarOpen, setSidebarOpen] = useState<boolean>(false);

  const handleToggleSidebar = () => {
    setSidebarOpen(!isSidebarOpen);
  }

  const closeSidebar = () => {
    setSidebarOpen(false);
  }

  return (
    <>
      <Header />
      <main className={styles.container}>
        <Breadcrumb />
        {/* TODO: Remove after testing; or implement a temporary button that toggles this in one of the pages */}
        <SidebarModal isOpen={isSidebarOpen} closeSidebar={closeSidebar} header='hi' body='hello' />
        <div className={styles.title}>
          <Title size={1}>{title}</Title>
          {button}
        </div>
        <LeadParagraph>{description}</LeadParagraph>
        {content}
        <button onClick={handleToggleSidebar}>Sidebar button test</button>
      </main>
    </>
  )
}
