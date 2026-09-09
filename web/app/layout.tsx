import type { ReactNode } from "react";
import type { Metadata, Viewport } from "next";
import { IBM_Plex_Sans, JetBrains_Mono, Syne } from "next/font/google";
import { siteDescription, siteName, siteTitle, siteUrl } from "../lib/site";
import { companyAddress, companyLegalName, sooappsUrl } from "../lib/legal";
import "./globals.css";

const sans = IBM_Plex_Sans({
  subsets: ["latin"],
  weight: ["400", "500", "600", "700"],
  variable: "--font-ibm-sans",
  display: "swap",
});

const mono = JetBrains_Mono({
  subsets: ["latin"],
  weight: ["400", "500"],
  variable: "--font-jetbrains-mono",
  display: "swap",
});

const wordmark = Syne({
  subsets: ["latin"],
  weight: ["600", "700"],
  variable: "--font-syne",
  display: "swap",
});

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
};

export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  title: { default: siteTitle, template: `%s — ${siteName}` },
  description: siteDescription,
  keywords: [
    "open source authentication",
    "self-hosted authentication",
    "OIDC provider",
    "OAuth provider",
    "passkey authentication",
    "Auth0 alternative",
    "Clerk alternative",
    "Sooauth",
    "Sooapps",
  ],
  authors: [{ name: "Sooapps", url: "https://sooapps.com" }],
  creator: "Sooapps",
  publisher: "Sooapps",
  category: "Developer tools",
  icons: {
    icon: [
      { url: "/favicon.ico", sizes: "any" },
      { url: "/icon-16.png", sizes: "16x16", type: "image/png" },
      { url: "/icon-32.png", sizes: "32x32", type: "image/png" },
      { url: "/icon-48.png", sizes: "48x48", type: "image/png" },
      { url: "/icon-192.png", sizes: "192x192", type: "image/png" },
      { url: "/icon-512.png", sizes: "512x512", type: "image/png" },
    ],
    apple: [
      { url: "/apple-touch-icon.png", sizes: "180x180", type: "image/png" },
    ],
    shortcut: "/favicon.ico",
  },
  robots: { index: true, follow: true },
  alternates: { canonical: "/" },
  openGraph: {
    type: "website",
    url: siteUrl,
    siteName,
    title: siteTitle,
    description: siteDescription,
  },
  twitter: {
    card: "summary_large_image",
    creator: "@sooapps",
    title: siteTitle,
    description: siteDescription,
  },
};

// Set the theme before first paint to avoid a flash.
const noFlashTheme = `(function(){try{var t=localStorage.getItem("theme");if(t==="dark"||t==="light"){document.documentElement.dataset.theme=t;}}catch(e){}})();`;
const organizationId = "https://sooapps.com/#organization";
const jsonLd = {
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "SoftwareApplication",
      "@id": `${siteUrl}/#software`,
      name: siteName,
      url: siteUrl,
      description: siteDescription,
      applicationCategory: "DeveloperApplication",
      operatingSystem: "Web",
      provider: { "@id": organizationId },
      creator: { "@id": organizationId },
      brand: { "@id": organizationId },
      offers: {
        "@type": "Offer",
        price: "0",
        priceCurrency: "USD",
        description: "Community Edition for self-hosting",
      },
    },
    {
      "@type": "Organization",
      "@id": organizationId,
      name: "Sooapps",
      legalName: companyLegalName,
      url: `${sooappsUrl}/`,
      email: "info@sooapps.com",
      address: {
        "@type": "PostalAddress",
        streetAddress: companyAddress,
        addressLocality: "Gebze",
        addressRegion: "Kocaeli",
        addressCountry: "TR",
      },
    },
  ],
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html
      lang="en"
      className={`${sans.variable} ${mono.variable} ${wordmark.variable}`}
    >
      <head>
        <script dangerouslySetInnerHTML={{ __html: noFlashTheme }} />
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
        />
      </head>
      <body className="bg-bg text-fg font-sans antialiased">{children}</body>
    </html>
  );
}
