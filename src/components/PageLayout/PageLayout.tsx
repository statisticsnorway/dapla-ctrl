import Breadcrumb from '../Breadcrumb'
import Header from '../Header/Header'
import SidebarModal from '../SidebarModal/SidebarModal'
import styles from './pagelayout.module.scss'

import { Title, LeadParagraph, Input, Dialog, Dropdown } from '@statisticsnorway/ssb-component-library'

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
        footer={{submitButtonText: "Legg til medlem"}}
        body={
          <div>
            <h2>Legg person til teamet</h2>
            <p>Navn</p>
            <Input placeholder="Skriv navn..." />

            <p>Rolle(r)</p>
            <Dropdown 
              placeholder="Velg rolle" 
              searchable
              selectedItem={{title: "Ingen tilgang", id: "ingen-tilgang"}}
              items={
              [
                {title: "Managers", id: "managers"},
                {title: "Developers", id: "developers"},
                {title: "Data-admins", id: "data-admins"},
                {title: "Ingen tilgang", id: "ingen-tilgang"},
                
              ]
            }/>

            <div className={styles.modalInfoDialog}></div>
            <Dialog type='info' title="Tidskrevende jobb!">
              Det kan ta litt tid før du ser endringen.
            </Dialog>
          </div>
        }
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
