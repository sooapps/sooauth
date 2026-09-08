import Link from "next/link";
import { ArrowRight, Check } from "lucide-react";
import { SiteFooter } from "./site-footer";
import { SiteHeader } from "./site-header";
import { buttonClass } from "./ui/button";
import { Container, Eyebrow } from "./ui/layout";
import { docsIntegrateUrl, signUpUrl } from "../lib/site";

type ComparisonPageProps = {
  competitor: string;
  description: string;
  bestFor: string;
  points: string[];
};

export function ComparisonPage({ competitor, description, bestFor, points }: ComparisonPageProps) {
  return (
    <>
      <SiteHeader />
      <main id="main">
        <Container>
          <article className="max-w-[900px] py-[clamp(3.5rem,7vw,7rem)]">
            <Eyebrow>Comparison guide</Eyebrow>
            <h1 className="mt-5 max-w-[760px] font-sans text-[clamp(2.5rem,5vw,4rem)] font-bold leading-[1.04] tracking-[-0.03em] text-fg">
              Sooauth vs {competitor}: choose the auth model you can live with.
            </h1>
            <p className="mt-6 max-w-[62ch] text-[18px] leading-[1.6] text-fg-muted">
              {description}
            </p>
            <div className="mt-9 flex flex-wrap gap-3">
              <a href={signUpUrl} className={buttonClass("primary")}>
                Try Sooauth free <ArrowRight size={16} aria-hidden />
              </a>
              <Link href={docsIntegrateUrl.replace("https://sooauth.com", "")} className={buttonClass("secondary")}>
                Read the OIDC guide
              </Link>
            </div>

            <section className="mt-16 grid gap-px border border-line bg-line md:grid-cols-2">
              <div className="bg-bg p-7 md:p-9">
                <p className="font-mono text-xs uppercase tracking-[0.14em] text-fg-muted">Sooauth</p>
                <h2 className="mt-4 font-sans text-[24px] font-semibold text-fg">Own the deployment choice</h2>
                <ul className="mt-6 grid gap-3 text-[15px] leading-[1.6] text-fg-muted">
                  {points.map((point) => (
                    <li key={point} className="flex gap-3">
                      <Check size={18} className="mt-0.5 shrink-0 text-fg" aria-hidden />
                      <span>{point}</span>
                    </li>
                  ))}
                </ul>
              </div>
              <div className="bg-bg-subtle p-7 md:p-9">
                <p className="font-mono text-xs uppercase tracking-[0.14em] text-fg-muted">Best fit</p>
                <h2 className="mt-4 font-sans text-[24px] font-semibold text-fg">{bestFor}</h2>
                <p className="mt-5 text-[15px] leading-[1.7] text-fg-muted">
                  Sooauth is a strong fit when you want a familiar hosted auth
                  experience today, standard OIDC at the application boundary,
                  and a clear self-hosting path for tomorrow.
                </p>
              </div>
            </section>

            <section className="mt-14 max-w-[680px]">
              <h2 className="font-sans text-[26px] font-semibold tracking-[-0.02em] text-fg">
                The important difference is portability
              </h2>
              <p className="mt-4 text-[16px] leading-[1.7] text-fg-muted">
                A hosted provider can be the fastest way to ship. Sooauth keeps
                that path while making deployment ownership a first-class choice.
                Your application integrates through OpenID Connect, so your
                product code does not need to know which deployment model runs
                behind the issuer URL.
              </p>
              <p className="mt-4 text-[16px] leading-[1.7] text-fg-muted">
                Compare the current limits, compliance requirements, support
                expectations, and total operating cost for your team. The right
                answer may be managed auth today and self-hosted auth later.
              </p>
            </section>
          </article>
        </Container>
      </main>
      <SiteFooter />
    </>
  );
}
