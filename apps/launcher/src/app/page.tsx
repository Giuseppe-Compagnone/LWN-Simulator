"use client";

import { Logo, LogoLayout, LogoSize } from "@lwn-simulator/ui-components";

export default function Home() {
  return (
    <div className={`home-page page`}>
      <Logo size={LogoSize.Lg} layout={LogoLayout.SecondaryBackground} />
      <h2 className="title">Establish Connection</h2>
      <h3 className="sub-title">Select the backend connection</h3>
    </div>
  );
}
