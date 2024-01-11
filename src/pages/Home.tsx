import { Title, Button } from '@statisticsnorway/ssb-component-library'

export function Home() {
  return (
    <div className="container">
      <div className="title-container">
        <Title size={1}>Teamoversikt</Title>
        <Button>Opprett team</Button>
      </div>
    </div>
  );
}
