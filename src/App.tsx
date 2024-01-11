import './App.scss'

import { Header } from './components/Header/Header'
import { Home } from './pages/Home'
import { Users } from './pages/Users'
import { Routes, Route } from 'react-router-dom'
import Breadcrumb from './components/Breadcrumb'

function App() {
  return (
    <main>
      <Header />
      <Breadcrumb />
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/medlemmer" element={<Users />} />
        <Route path="/medlemmer/test" element={<h1>Test</h1>} />
      </Routes>
    </main >
  )
}

export default App
