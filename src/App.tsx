import './App.scss';

import { Header } from './components/Header/Header';
import { ProtectedRoute } from './components/ProtectedRoute';
import { Home } from './pages/Home';
import { Users } from './pages/Users';
import { Routes, Route } from 'react-router-dom';
import Breadcrumb from './components/Breadcrumb';
import { Login } from './pages/Login';

function App() {
  return (
    <main>
      <Header />
      <Breadcrumb />
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/login" element={<Login />} />

        <Route element={<ProtectedRoute />}>
          <Route path="/medlemmer" element={<Users />} />
          <Route path="/medlemmer/test" element={<h1>Test</h1>} />
        </Route>
      </Routes>
    </main>
  )
}

export default App;
