import './App.scss'

import { Header } from './components/Header/Header'
import { Home } from './pages/Home'
import { Users } from './pages/Users'
import { Routes, Route } from 'react-router-dom'

function App() {
  return (
    <main>
      <Header />
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/users" element={<Users />} />
      </Routes>
    </main>
  )
}

export default App
