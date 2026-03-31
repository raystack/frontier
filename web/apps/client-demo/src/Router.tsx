import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Home from './pages/Home';
import Login from './pages/Login';
import Signup from './pages/Signup';
import Callback from './pages/Callback';
import MagiclinkVerify from './pages/MagiclinkVerify';
import Subscribe from './pages/Subscribe';
import Updates from './pages/Updates';
import Organization from './pages/Organization';

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
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default Router;