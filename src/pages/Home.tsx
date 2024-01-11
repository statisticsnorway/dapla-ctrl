import { Breadcrumb, Title, Button } from '@statisticsnorway/ssb-component-library'

export function Home() {
    const breadcrumbItems = [{text: 'Forside'}]
  
    return (
      <div className="container">
        <Breadcrumb items={breadcrumbItems}/>
        <Title size={1}>Teamoversikt</Title>
        <Button>Opprett team</Button>
      </div>
    )
}