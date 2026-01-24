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

```
NEXT_PUBLIC_API_URL=http://localhost:8000
NEXT_PUBLIC_DOMAIN=localhost
```

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.

