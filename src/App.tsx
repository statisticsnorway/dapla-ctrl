import styles from './app.module.scss'

import Header from './components/Header/Header'
import Breadcrumb from './components/Breadcrumb'
import Home from './pages/Home'
import Users from './pages/Users'

import { Routes, Route } from 'react-router-dom'

export default function App() {
  return (
    <>
      <Header />
      <main className={styles.container}>
      <Breadcrumb />
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/medlemmer" element={<Users />} />
          <Route path="/medlemmer/test" element={<h1>Test</h1>} />
        </Routes>
      </main >
    </>
  )
}