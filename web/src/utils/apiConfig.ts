import { OpenAPI } from "../api";

export function initApi() {
  OpenAPI.BASE = import.meta.env.VITE_API_BASE || "";
}