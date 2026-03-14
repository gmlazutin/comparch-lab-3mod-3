import React, { useState } from "react";
import { useNavigate } from "react-router";
import { useAuth } from "../auth/AuthContext";
import { DefaultService } from "../api";
import { apiCall } from "../utils/apiCall";

const LoginPage: React.FC = () => {
    const { setSession } = useAuth();
    const navigate = useNavigate();
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");

    const handleLogin = async () => {
        const sess = await apiCall(() => 
            DefaultService.authUser({
                login: username,
                password: password
            })
        );
        if (!sess.ok) {
            return
        }
        setSession({
            token: sess.data.auth.token,
            expires: new Date(sess.data.auth.expires).getTime()
        });
    };

    return (
        <div className="container mt-5">
            <h2>Login</h2>
            <div className="mb-3">
                <label className="form-label">Username</label>
                <input className="form-control" value={username} onChange={e => setUsername(e.target.value)} />
            </div>
            <div className="mb-3">
                <label className="form-label">Password</label>
                <input type="password" className="form-control" value={password} onChange={e => setPassword(e.target.value)} />
            </div>
            <button className="btn btn-primary me-2" onClick={handleLogin}>Login</button>
            <button className="btn btn-link" onClick={() => navigate("/register")}>Register</button>
        </div>
    );
};

export default LoginPage;