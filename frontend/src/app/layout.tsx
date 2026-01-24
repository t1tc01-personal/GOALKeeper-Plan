import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'
import { Toaster } from '@/components/ui/sonner'
import QueryProvider from '@/shared/contexts/QueryProvider'
import AuthProvider from '@/shared/contexts/AuthProvider'
import { getAuthToken } from '@/shared/lib/authCookie'
import NextTopLoader from 'nextjs-toploader'

const inter = Inter({
  variable: '--font-inter',
  subsets: ['latin'],
})

export const metadata: Metadata = {
  title: 'GOALKeeper Plan',
  description: 'GOALKeeper Plan is a platform for goal management and planning',
}

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  const token = await getAuthToken()
  const isAuthenticated = !!token
  return (
    <html lang="en">
      <body className={`${inter.variable} antialiased`}>
        <QueryProvider>
          <AuthProvider isAuthenticated={isAuthenticated}>
            {children}
          </AuthProvider>
          <NextTopLoader color="#5163FF" showSpinner={false} />
          <Toaster position="top-right" />
        </QueryProvider>
      </body>
    </html>
  )
}

