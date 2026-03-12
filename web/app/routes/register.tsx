import { useState } from "react";
import { useNavigate, Link } from "react-router";
import { DefaultService } from "../api";
import { useAuth } from "../context/AuthContext";

export default function Register() {
  const [login, setLogin] = useState("");
  const [password, setPassword] = useState("");

  const { setAuth } = useAuth();
  const nav = useNavigate();

  async function submit() {
    try {
      const res = await DefaultService.postApiV1AuthRegister({
        login,
        password,
      });

      const expiresTs = new Date(res.auth.expires).getTime();
      setAuth(res.auth.token, expiresTs);

      nav("/");
    } catch (e: any) {
      alert(e.body?.error);
    }
  }

  return (
    <div>
      <h1>Register</h1>

      <input
        placeholder="login"
        value={login}
        onChange={(e) => setLogin(e.target.value)}
      />

      <input
        type="password"
        placeholder="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
      />

      <button onClick={submit}>register</button>

      <Link to="/login">login</Link>
    </div>
  );
}