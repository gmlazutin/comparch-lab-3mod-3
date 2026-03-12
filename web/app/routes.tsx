import type { RouteConfig } from "@react-router/dev/routes";

export default [
  { path: "/login", file: "./routes/login.tsx" },
  { path: "/register", file: "./routes/register.tsx" },
  { path: "/", file: "./routes/contacts.tsx" }
] satisfies RouteConfig;