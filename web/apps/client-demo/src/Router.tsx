import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Home from './pages/Home';
import Login from './pages/Login';
import Signup from './pages/Signup';
import Callback from './pages/Callback';
import MagiclinkVerify from './pages/MagiclinkVerify';
import Subscribe from './pages/Subscribe';
import Updates from './pages/Updates';
import Organization from './pages/Organization';
import Settings from './pages/Settings';
import General from './pages/settings/General';
import Preferences from './pages/settings/Preferences';
import Profile from './pages/settings/Profile';
import Sessions from './pages/settings/Sessions';
import Members from './pages/settings/Members';

function Router() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/login" element={<Login />} />
        <Route path="/signup" element={<Signup />} />
        <Route path="/callback" element={<Callback />} />
        <Route path="/magiclink-verify" element={<MagiclinkVerify />} />
        <Route path="/subscribe" element={<Subscribe />} />
        <Route path="/updates" element={<Updates />} />
        <Route path="/organizations/:orgId" element={<Organization />} />
        <Route path="/:orgId/settings" element={<Settings />}>
          <Route path="general" element={<General />} />
          <Route path="preferences" element={<Preferences />} />
          <Route path="profile" element={<Profile />} />
          <Route path="sessions" element={<Sessions />} />
          <Route path="members" element={<Members />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default Router;