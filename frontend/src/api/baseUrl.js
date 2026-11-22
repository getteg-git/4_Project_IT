const configuredBaseUrl = import.meta.env.VITE_API_BASE_URL?.trim();

export const API_BASE_URL = (
  configuredBaseUrl ||
  "http://localhost:8080"
).replace(/\/+$/, "");

export const apiFetch = (input, init = {}) =>
  fetch(input, { ...init, credentials: "include" });
