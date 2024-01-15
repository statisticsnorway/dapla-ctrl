import styles from './app.module.scss'

import Header from './components/Header/Header';
import Breadcrumb from './components/Breadcrumb';
import { ProtectedRoute } from './components/ProtectedRoute';
import Home from './pages/Home';
import Users from './pages/Users';
import Login from './pages/Login';

import { Routes, Route } from 'react-router-dom';

export default function App() {
  return (
    <>
      <Header />
      <main className={styles.container}>
        <Breadcrumb />
        <Routes>
          <Route path="/login" element={<Login />} />

          <Route element={<ProtectedRoute />}> {
            /* Possibly setup passable props to ProtectedRoute so we can add authorization too,
            example: <ProtectedRoute roles={['managers', 'data-admins']} />
            */
          }
            <Route path="/" element={<Home />} />
            <Route path="/medlemmer" element={<Users />} />
            <Route path="/medlemmer/test" element={<h1>Test</h1>} />
          </Route>
        </Routes>
      </main>
    </>
  )
}
