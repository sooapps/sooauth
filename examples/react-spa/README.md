# React example

One Vite app that signs in through sooauth three ways: the ready-made hosted
page (OIDC with PKCE) at `/`, the drop-in widget at `/widget`, and your own forms
over the API at `/login`.

1. Start a local sooauth with captured email:

   ```bash
   docker compose -f docker-compose.test.yml -p sooauth-test up -d --build
   ```

2. Sign up at <http://localhost:8090/auth/sign-up>. Emails land in <http://localhost:8091>.
3. In your project, open **Keys & URLs** and add `http://localhost:5173/callback`
   and `http://localhost:5173/login` as redirect URIs.
4. Copy `.env.example` to `.env.local` and set your client ID.
5. Run `pnpm install --ignore-workspace && pnpm dev` here and open <http://localhost:5173>.
