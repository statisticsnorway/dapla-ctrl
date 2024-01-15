import styles from './app.module.scss'

import Header from './components/Header/Header';
import Breadcrumb from './components/Breadcrumb';
import { ProtectedRoute } from './components/ProtectedRoute';
import Home from './pages/Home';
import Users from './pages/Users';
import Logout from './pages/Logout';
import Login from './pages/Login/Login';

import { Routes, Route, useLocation } from 'react-router-dom';

export default function App() {
  const isLoggedIn = useLocation().pathname !== '/login';

  return (
    <>
      <Header isLoggedIn={isLoggedIn} />
      <main className={styles.container}>
        {isLoggedIn && <Breadcrumb />}
        <Routes>
          <Route path="/login" element={<Login />} />

          <Route element={<ProtectedRoute />}> {
            /* Possibly setup passable props to ProtectedRoute so we can add authorization too,
            example: <ProtectedRoute roles={['managers', 'data-admins']} />
            */
          }
            <Route path="/" element={<Home />} />
            <Route path="/logout" element={<Logout />} />
            <Route path="/teammedlemmer" element={<Users />} />
            <Route path="/teammedlemmer/:user" element={<h1>Test</h1>} />
          </Route>
        </Routes>
      </main>
    </>
  )
}
