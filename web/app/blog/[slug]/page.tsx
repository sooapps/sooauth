import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { SiteFooter } from "../../../components/site-footer";
import { SiteHeader } from "../../../components/site-header";
import { Container, Eyebrow } from "../../../components/ui/layout";
import { blogPosts, getBlogPost } from "../../../lib/blog";
import { siteUrl } from "../../../lib/site";

export function generateStaticParams() {
  return blogPosts.map((post) => ({ slug: post.slug }));
}

export async function generateMetadata({ params }: { params: Promise<{ slug: string }> }): Promise<Metadata> {
  const post = getBlogPost((await params).slug);
  if (!post) return {};
  return {
    title: post.title,
    description: post.description,
    alternates: { canonical: `/blog/${post.slug}` },
    openGraph: { type: "article", publishedTime: post.date },
  };
}

export default async function BlogPostPage({ params }: { params: Promise<{ slug: string }> }) {
  const post = getBlogPost((await params).slug);
  if (!post) notFound();

  const articleJsonLd = {
    "@context": "https://schema.org",
    "@type": "Article",
    headline: post.title,
    description: post.description,
    datePublished: post.date,
    dateModified: post.date,
    author: { "@type": "Organization", name: "Sooapps", url: "https://sooapps.com" },
    publisher: { "@type": "Organization", name: "Sooapps", url: "https://sooapps.com" },
    mainEntityOfPage: `${siteUrl}/blog/${post.slug}`,
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(articleJsonLd) }}
      />
      <SiteHeader />
      <main id="main">
        <Container>
          <article className="mx-auto max-w-[720px] py-[clamp(3.5rem,7vw,7rem)]">
            <Link href="/blog" className="font-mono text-xs uppercase tracking-[0.12em] text-fg-muted hover:text-fg">
              ← Back to blog
            </Link>
            <div className="mt-10">
              <Eyebrow>{post.eyebrow} · {post.readTime}</Eyebrow>
              <h1 className="mt-5 font-sans text-[clamp(2.25rem,5vw,3.75rem)] font-bold leading-[1.05] tracking-[-0.03em] text-fg">
                {post.title}
              </h1>
              <p className="mt-6 text-[18px] leading-[1.6] text-fg-muted">
                {post.description}
              </p>
              <p className="mt-5 font-mono text-xs uppercase tracking-[0.12em] text-fg-muted">
                Published {post.date}
              </p>
            </div>
            <div className="mt-14 grid gap-10">
              {post.sections.map((section) => (
                <section key={section.heading}>
                  <h2 className="font-sans text-[24px] font-semibold tracking-[-0.02em] text-fg">
                    {section.heading}
                  </h2>
                  <div className="mt-4 grid gap-4 text-[16px] leading-[1.7] text-fg-muted">
                    {section.paragraphs.map((paragraph) => (
                      <p key={paragraph}>{paragraph}</p>
                    ))}
                  </div>
                  {section.bullets ? (
                    <ul className="mt-5 grid gap-2 border-l-2 border-line-strong pl-5 text-[15px] leading-[1.6] text-fg-muted">
                      {section.bullets.map((bullet) => <li key={bullet}>{bullet}</li>)}
                    </ul>
                  ) : null}
                </section>
              ))}
            </div>
            <div className="mt-14 border border-line bg-bg-subtle p-6 md:p-8">
              <h2 className="font-sans text-[21px] font-semibold text-fg">Try Sooauth</h2>
              <p className="mt-2 text-[15px] leading-[1.6] text-fg-muted">
                Start with hosted auth or follow the self-host guide. Your app uses standard OIDC either way.
              </p>
              <Link href="/docs/getting-started" className="mt-5 inline-flex font-mono text-xs uppercase tracking-[0.12em] text-fg underline underline-offset-4">
                Read the getting started guide →
              </Link>
            </div>
          </article>
        </Container>
      </main>
      <SiteFooter />
    </>
  );
}
