import Image from "next/image";

export function BrandLogo({ className = "" }: { className?: string }) {
  return (
    <span className={`brand-logo ${className}`}>
      <Image src="/sooauth-light.png" alt="sooauth" width={36} height={36} className="brand-logo-image brand-logo-light" priority />
      <Image src="/sooauth-dark.png" alt="" width={36} height={36} className="brand-logo-image brand-logo-dark" priority />
      <span className="font-wordmark text-[18px] font-bold tracking-[-0.03em] text-fg">
        sooauth
      </span>
    </span>
  );
}
