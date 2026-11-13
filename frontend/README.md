# React + Vite

## Local development

1. Start PostgreSQL from the repository root:

   ```sh
   docker compose -f database/docker-compose.yml up -d db
   ```

2. Make sure `backend/.env` contains local PostgreSQL settings and SMTP settings if email notifications are needed. `backend/.env.example` is the template; do not overwrite an existing `.env` that contains your local credentials.
3. Start the API from the `backend` directory:

   ```sh
   go run main.go
   ```

4. Start the frontend from the `frontend` directory:

   ```sh
   npm run dev
   ```

   In Vite development mode, the API defaults to `http://localhost:8080`. `frontend/.env` is optional; if used, set `VITE_API_BASE_URL=http://localhost:8080`.

This template provides a minimal setup to get React working in Vite with HMR and some ESLint rules.

Currently, two official plugins are available:

- [@vitejs/plugin-react](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react) uses [Babel](https://babeljs.io/) (or [oxc](https://oxc.rs) when used in [rolldown-vite](https://vite.dev/guide/rolldown)) for Fast Refresh
- [@vitejs/plugin-react-swc](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react-swc) uses [SWC](https://swc.rs/) for Fast Refresh

## React Compiler

The React Compiler is not enabled on this template because of its impact on dev & build performances. To add it, see [this documentation](https://react.dev/learn/react-compiler/installation).

## Expanding the ESLint configuration

If you are developing a production application, we recommend using TypeScript with type-aware lint rules enabled. Check out the [TS template](https://github.com/vitejs/vite/tree/main/packages/create-vite/template-react-ts) for information on how to integrate TypeScript and [`typescript-eslint`](https://typescript-eslint.io) in your project.
