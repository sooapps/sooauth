import type { MetadataRoute } from "next";
import { blogPosts } from "../lib/blog";
import { siteUrl } from "../lib/site";

export default function sitemap(): MetadataRoute.Sitemap {
  const routes = [
    "",
    "/docs",
    "/docs/getting-started",
    "/docs/integrate-oidc",
    "/docs/custom-auth-ui",
    "/docs/embed-widget",
    "/docs/google-oauth",
    "/docs/admin",
    "/docs/self-host",
    "/blog",
    "/compare/auth0",
    "/compare/better-auth",
    "/compare/clerk",
  ];

  return [
    ...routes.map((route) => ({
      url: `${siteUrl}${route}`,
      changeFrequency: route === "" ? "weekly" as const : "monthly" as const,
      priority: route === "" ? 1 : 0.7,
    })),
    ...blogPosts.map((post) => ({
      url: `${siteUrl}/blog/${post.slug}`,
      changeFrequency: "monthly" as const,
      priority: 0.6,
      lastModified: post.date,
    })),
  ];
}
