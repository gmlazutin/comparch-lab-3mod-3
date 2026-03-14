import React from "react";
import { Routes, Route, Navigate } from "react-router";
import LoginPage from "./pages/LoginPage";
import RegisterPage from "./pages/RegisterPage";
import ContactsPage from "./pages/ContactsPage";
import { useAuth } from "./auth/AuthContext";
import { initApi } from "./utils/apiConfig"

initApi();

const App: React.FC = () => {
    const { session } = useAuth();

    return (
        <Routes>
            <Route path="/" element={session ? <ContactsPage /> : <Navigate to="/login" />} />
            <Route path="/login" element={session ? <Navigate to="/" /> : <LoginPage />} />
            <Route path="/register" element={session ? <Navigate to="/" /> : <RegisterPage />} />
            <Route path="*" element={<Navigate to="/" />} />
        </Routes>
    );
};

export default App;