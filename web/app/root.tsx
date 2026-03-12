import { Outlet } from "react-router";
import { AuthProvider } from "./context/AuthContext";
import { initApi } from "./apiConfig";

initApi();

export default function Root() {
  return (
    <AuthProvider>
      <Outlet />
    </AuthProvider>
  );
}