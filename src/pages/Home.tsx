import { Layout } from "../components/Layout/Layout"

export function Home() {  
    return (
      <Layout 
        title="Teamoversikt" 
        breadcrumbItems={[{text: 'Forside'}]}
        buttonText="Opprett team"
      />
    )
}