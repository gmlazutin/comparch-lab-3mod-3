import { OpenAPI } from "../api";

export function initApi() {
  OpenAPI.BASE = import.meta.env.API_BASE || "/";
}