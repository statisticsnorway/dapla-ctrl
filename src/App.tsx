import ProtectedRoute from './components/ProtectedRoute'
import ProtectedAuthorizedUserRoute from './components/ProtectedAuthorizedUserRoute'

import CreateTeamForm from './pages/CreateTeamForm/CreateTeamForm.tsx'
import TeamCreated from './pages/CreateTeamForm/TeamCreated.tsx'
import NotFound from './pages/NotFound/NotFound.tsx'
import TeamOverview from './pages/TeamOverview/TeamOverview'
import UserProfile from './pages/UserProfile/UserProfile'
import TeamDetail from './pages/TeamDetail/TeamDetail'
import TeamMembers from './pages/TeamMembers/TeamMembers'
import SharedBucketDetail from './pages/SharedBucketDetail/SharedBucketDetail.tsx'

import { Routes, Route } from 'react-router-dom'

const App = () => {
  return (
    <Routes>
      <Route element={<ProtectedRoute />}>
        <Route path='/' element={<TeamOverview />} />
        <Route path='/teammedlemmer' element={<TeamMembers />} />
        <Route path='/teammedlemmer/:principalName' element={<UserProfile />} />
        <Route path='/:teamId' element={<TeamDetail />} />
        <Route path='/:teamId/:shortName' element={<SharedBucketDetail />} />
        <Route path='/opprett-team' element={<ProtectedAuthorizedUserRoute component={<CreateTeamForm />} />} />
        <Route path='/opprett-team/kvittering' element={<ProtectedAuthorizedUserRoute component={<TeamCreated />} />} />
        <Route path='/not-found' element={<NotFound />} />
      </Route>
    </Routes>
  )
}

export default App
