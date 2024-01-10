import './styles/App.scss'
import { Breadcrumb } from '@statisticsnorway/ssb-component-library'

function App() {
  return (
    <>
      <Breadcrumb items={[{text: 'Forside'}]} />
      <h1>Teamoversikt</h1>
    </>
  )
}

export default App
