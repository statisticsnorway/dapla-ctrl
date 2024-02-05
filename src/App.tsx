import ProtectedRoute from './components/ProtectedRoute'

import NotFound from './pages/NotFound/NotFound.tsx'
import Login from './pages/Login/Login'
import TeamOverview from './pages/TeamOverview/TeamOverview'
import UserProfile from './pages/UserProfile/UserProfile'
import TeamDetail from './pages/TeamDetail/TeamDetail'

import { Routes, Route } from 'react-router-dom'

export default function App() {
  return (
    <Routes>
      {import.meta.env.NODE_ENV !== 'production' && <Route path='/login' element={<Login />} />}
      <Route element={<ProtectedRoute />}>
        {/* Possibly setup passable props to ProtectedRoute so we can add authorization too,
            example: <ProtectedRoute roles={['managers', 'data-admins']} />
            */}
        <Route path='/' element={<TeamOverview />} />
        <Route path='/teammedlemmer' element={<h1>Teammedlemmer</h1>} />
        <Route path='/teammedlemmer/:principalName' element={<UserProfile />} />
        <Route path='/:teamId' element={<TeamDetail />} />
        <Route path='*' element={<NotFound />} />
      </Route>
    </Routes>
  )
}
