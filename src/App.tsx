import ProtectedRoute from './components/ProtectedRoute'

import NotFound from './pages/NotFound/NotFound.tsx'
import TeamOverview from './pages/TeamOverview/TeamOverview'
import UserProfile from './pages/UserProfile/UserProfile'
import TeamDetail from './pages/TeamDetail/TeamDetail'
import TeamMembers from './pages/TeamMembers/TeamMembers'

import { Routes, Route } from 'react-router-dom'

const App = () => {
  return (
    <Routes>
      <Route element={<ProtectedRoute />}>
        <Route path='/' element={<TeamOverview />} />
        <Route path='/teammedlemmer' element={<TeamMembers />} />
        <Route path='/teammedlemmer/:principalName' element={<UserProfile />} />
        <Route path='/:teamId' element={<TeamDetail />} />
        <Route path='*' element={<NotFound />} />
      </Route>
    </Routes>
  )
}

export default App
