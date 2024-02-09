import { useState } from 'react'
import Breadcrumb from '../Breadcrumb'
import Header from '../Header/Header'
import SidebarModal from '../SidebarModal/SidebarModal'
import styles from './pagelayout.module.scss'

import { Title, LeadParagraph, Input, Dropdown, Dialog, Button } from '@statisticsnorway/ssb-component-library'

interface PageLayoutProps {
  title: string
  description?: JSX.Element
  button?: JSX.Element
  content?: JSX.Element
}

export default function PageLayout({ title, description, button, content }: PageLayoutProps) {
  const [open, setOpen] = useState<boolean>(false)

  const modalBody = (): JSX.Element => {
    return (
      <div>
        <h2>Legg person til teamet</h2>
        <p>Navn</p>
        <Input placeholder='Skriv navn...' />

        <p>Tilgangsgruppe(r)</p>
        <Dropdown
          placeholder='Velg gruppe(r)'
          searchable
          selectedItem={{ title: 'Ingen tilgang', id: 'ingen-tilgang' }}
          items={[
            { title: 'Managers', id: 'managers' },
            { title: 'Developers', id: 'developers' },
            { title: 'Data-admins', id: 'data-admins' },
            { title: 'Ingen tilgang', id: 'ingen-tilgang' },
          ]}
        />

        <div className={styles.modalInfoDialog}>
          <Dialog type='info' title='Tidskrevende jobb!'>
            Det kan ta litt tid før du ser endringen.
          </Dialog>
        </div>
      </div>
    )
  }

  const handleModalClose = () => {
    setOpen(false)
  }

  return (
    <>
      <Header />
      <main className={styles.container}>
        <Breadcrumb />
        {/* TODO: Remove after testing; or implement a temporary button that toggles this in one of the pages */}
        <Button onClick={() => setOpen(true)}>Click me</Button>
        <SidebarModal
          open={open}
          onClose={handleModalClose}
          header={{ modalType: 'Medlem', modalTitle: 'Arbmark register', modalDescription: 'arbmark-register' }}
          footer={{
            submitButtonText: 'Legg til medlem',
            handleSubmit: () => {
              console.log('clicked')
              setOpen(false)
            },
          }}
          body={modalBody()}
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
