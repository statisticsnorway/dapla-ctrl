import styles from './app.module.scss'

import Header from './components/Header/Header';
import Breadcrumb from './components/Breadcrumb';
import ProtectedRoute from './components/ProtectedRoute';

import Login from './pages/Login/Login';
import TeamOverview from './pages/TeamOverview/TeamOverview';
import UserProfile from './pages/UserProfile/UserProfile';

import { Routes, Route } from 'react-router-dom';
import { useContext } from 'react';
import { DaplaCtrlContext } from './provider/DaplaCtrlProvider';

export default function App() {
  const { isLoggedIn } = useContext(DaplaCtrlContext);

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
            <Route path="/" element={<TeamOverview />} />
            <Route path="/teammedlemmer" element={<h1>Teammedlemmer</h1>} />
            <Route path="/teammedlemmer/:principalName" element={<UserProfile />} />
          </Route>
        </Routes>
      </main>
    </>
  )
}
