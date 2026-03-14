import React, { useState } from "react";
import { useNavigate } from "react-router";
import { DefaultService } from "../api";
import { useAuth } from "../auth/AuthContext";
import { apiCall } from "../utils/apiCall";

const RegisterPage: React.FC = () => {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [password2, setPassword2] = useState("");
    const { setSession } = useAuth();
    const navigate = useNavigate();

    const handleRegister = async () => {
        if (password !== password2) {
            alert("Passwords do not match");
            return;
        }
        const sess = await apiCall(() =>
            DefaultService.registerUser({
                login: username,
                password: password
            })
        );
        if (!sess.ok) {
            return;
        }
        setSession({
            token: sess.data.auth.token,
            expires: new Date(sess.data.auth.expires).getTime()
        });
    };

    return (
        <div className="container mt-5">
            <h2>Register</h2>
            <div className="mb-3">
                <label className="form-label">Username</label>
                <input className="form-control" value={username} onChange={e => setUsername(e.target.value)} />
            </div>
            <div className="mb-3">
                <label className="form-label">Password</label>
                <input type="password" className="form-control" value={password} onChange={e => setPassword(e.target.value)} />
            </div>
            <div className="mb-3">
                <label className="form-label">Repeat Password</label>
                <input type="password" className="form-control" value={password2} onChange={e => setPassword2(e.target.value)} />
            </div>
            <button className="btn btn-primary me-2" onClick={handleRegister}>Register</button>
            <button className="btn btn-link" onClick={() => navigate("/login")}>Login</button>
        </div>
    );
};

export default RegisterPage;