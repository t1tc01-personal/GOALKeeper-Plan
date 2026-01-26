This is a [Next.js](https://nextjs.org) project bootstrapped with [`create-next-app`](https://nextjs.org/docs/app/api-reference/cli/create-next-app).

## Getting Started

First, run the development server:

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
# or
bun dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

You can start editing the page by modifying `app/page.tsx`. The page auto-updates as you edit the file.

## Project Structure

The frontend follows a clean architecture pattern similar to learninghub-frontend:

```
frontend/
├── src/
│   ├── app/              # Next.js App Router pages and layouts
│   ├── components/       # Reusable UI components
│   │   └── ui/          # Shadcn UI components
│   ├── config/          # Configuration files
│   ├── features/        # Feature-based modules
│   ├── shared/          # Shared utilities, types, contexts
│   │   ├── constant/   # Constants and enums
│   │   ├── contexts/   # React contexts (Auth, Query, etc.)
│   │   ├── hooks/      # Custom React hooks
│   │   ├── lib/        # Core libraries (httpClient, middleware, etc.)
│   │   ├── types/      # TypeScript types
│   │   └── utils/      # Utility functions
│   └── proxy.ts        # Middleware proxy for authentication
├── package.json
├── tsconfig.json
├── next.config.ts
└── postcss.config.mjs
```

## Features

- ✅ Next.js 16 with App Router
- ✅ TypeScript with strict mode
- ✅ Tailwind CSS v4
- ✅ React Query for data fetching
- ✅ Authentication middleware
- ✅ HTTP client with automatic token refresh
- ✅ Shadcn UI components
- ✅ Radix UI primitives
- ✅ Sonner for toast notifications

## Environment Variables

Create a `.env.local` file in the root directory:

```env
# API Configuration
NEXT_PUBLIC_API_URL=http://localhost:8000
NEXT_PUBLIC_DOMAIN=http://localhost:3000

# OAuth Client IDs (used in frontend - can use NEXT_PUBLIC_* prefix or without)
GITHUB_CLIENT_ID=your_github_client_id
GOOGLE_CLIENT_ID=your_google_client_id

# OAuth Client Secrets (server-side only, for token exchange)
GITHUB_CLIENT_SECRET=your_github_client_secret
GOOGLE_CLIENT_SECRET=your_google_client_secret
```

**Important:** 
- `NEXT_PUBLIC_DOMAIN` should be set to `http://localhost:3000` (full URL with protocol)
- `GITHUB_CLIENT_ID` and `GOOGLE_CLIENT_ID` are used in frontend to generate OAuth URLs
  - You can use `NEXT_PUBLIC_GITHUB_CLIENT_ID` and `NEXT_PUBLIC_GOOGLE_CLIENT_ID` instead if preferred
  - The code supports both naming conventions
- `GITHUB_CLIENT_SECRET` and `GOOGLE_CLIENT_SECRET` are only used in Next.js API routes (server-side)
- OAuth redirect URIs must be configured in your OAuth provider apps:
  - GitHub: `http://localhost:3000/auth/github/callback`
  - Google: `http://localhost:3000/auth/google/callback`

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.

