import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'
import { Toaster } from '@/components/ui/sonner'
import QueryProvider from '@/shared/contexts/QueryProvider'
import AuthProvider from '@/shared/contexts/AuthProvider'
import NextTopLoader from 'nextjs-toploader'

const inter = Inter({
  variable: '--font-inter',
  subsets: ['latin'],
})

export const metadata: Metadata = {
  title: 'GOALKeeper Plan',
  description: 'GOALKeeper Plan is a platform for goal management and planning',
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en">
      <body className={`${inter.variable} antialiased`}>
        <QueryProvider>
          <AuthProvider>
            {children}
          </AuthProvider>
          <NextTopLoader color="#5163FF" showSpinner={false} />
          <Toaster position="top-right" />
        </QueryProvider>
      </body>
    </html>
  )
}

