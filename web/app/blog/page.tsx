import type { Metadata } from "next";
import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { SiteFooter } from "../../components/site-footer";
import { SiteHeader } from "../../components/site-header";
import { Container, Eyebrow } from "../../components/ui/layout";
import { blogPosts } from "../../lib/blog";

export const metadata: Metadata = {
  title: "Blog",
  description:
    "Product, engineering, and deployment notes from the Sooapps team building Sooauth.",
  alternates: { canonical: "/blog" },
};

export default function BlogPage() {
  return (
    <>
      <SiteHeader />
      <main id="main">
        <Container>
          <div className="max-w-[680px] py-[clamp(3.5rem,7vw,7rem)]">
            <Eyebrow>From the Sooapps team</Eyebrow>
            <h1 className="mt-5 font-sans text-[clamp(2.5rem,5vw,4rem)] font-bold leading-[1.04] tracking-[-0.03em] text-fg">
              Notes on auth you can own.
            </h1>
            <p className="mt-6 text-[17px] leading-[1.6] text-fg-muted">
              Product decisions, OIDC engineering, and practical deployment
              guidance for teams building SaaS products.
            </p>
          </div>
          <div className="grid gap-4 md:grid-cols-3">
            {blogPosts.map((post) => (
              <article key={post.slug} className="flex flex-col border border-line bg-bg p-6 md:p-8">
                <p className="font-mono text-xs uppercase tracking-[0.14em] text-fg-muted">
                  {post.eyebrow} · {post.readTime}
                </p>
                <h2 className="mt-5 font-sans text-[22px] font-semibold leading-[1.2] tracking-[-0.01em] text-fg">
                  <Link href={`/blog/${post.slug}`} className="hover:underline">
                    {post.title}
                  </Link>
                </h2>
                <p className="mt-4 flex-1 text-[15px] leading-[1.6] text-fg-muted">
                  {post.description}
                </p>
                <Link
                  href={`/blog/${post.slug}`}
                  className="mt-8 inline-flex items-center gap-2 font-mono text-xs uppercase tracking-[0.12em] text-fg"
                >
                  Read article <ArrowRight size={14} aria-hidden />
                </Link>
              </article>
            ))}
          </div>
        </Container>
      </main>
      <SiteFooter />
    </>
  );
}
