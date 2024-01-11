import { PageLayout } from "../components/PageLayout/PageLayout"

export function Home() {  
    return (
      <PageLayout 
        title="Teamoversikt" 
        breadcrumbItems={[{text: 'Forside'}]}
        buttonText="Opprett team"
      />
    )
}