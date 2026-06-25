import { Routes, Route } from "react-router-dom";
import Layout from "./components/Layout";
import {
  Landing,
  Login,
  Register,
  Onboarding,
  ProfilePage,
  OfficeHours,
  OfficeHoursSession,
  DataRoom,
  Learn,
  LearnSession,
  Goals,
  Dashboard,
  Logbook,
} from "./pages/pages";

// All routes from BUILD_SPEC §7.
export default function App() {
  return (
    <Routes>
      {/* Unauthenticated / standalone pages (no sidebar). */}
      <Route path="/" element={<Landing />} />
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />

      {/* App pages share the sidebar layout. */}
      <Route element={<Layout />}>
        <Route path="/onboarding" element={<Onboarding />} />
        <Route path="/profile" element={<ProfilePage />} />
        <Route path="/office-hours" element={<OfficeHours />} />
        <Route path="/office-hours/:id" element={<OfficeHoursSession />} />
        <Route path="/data-room" element={<DataRoom />} />
        <Route path="/learn" element={<Learn />} />
        <Route path="/learn/:key" element={<LearnSession />} />
        <Route path="/goals" element={<Goals />} />
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/logbook" element={<Logbook />} />
      </Route>
    </Routes>
  );
}
