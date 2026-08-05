// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore
import "./globals.css";
import { Footer, Layout, Navbar } from "nextra-theme-docs";
import { Head } from "nextra/components";
import { getPageMap } from "nextra/page-map";
import type { Metadata } from "next";
import type { ReactNode } from "react";
import Image from "next/image";

export const metadata: Metadata = {
  title: "LWN Simulator Docs",
  description: "Documentation of LWN Simulator",
  icons: ["/favicon.png"],
};

const footer = <Footer>MIT {new Date().getFullYear()} © Nextra.</Footer>;

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  return (
    <html lang="en" dir="ltr" suppressHydrationWarning>
      <Head />
      <body>
        <Layout
          navbar={
            <Navbar
              logo={
                <div className="flex items-center gap-2">
                  <Image
                    src="/favicon.png"
                    alt="logo"
                    width={50}
                    height={50}
                    className="w-12 h-12"
                  />
                  <h1 className="text-2xl">LWN Simulator</h1>
                </div>
              }
              projectLink="https://github.com/Giuseppe-Compagnone/LWN-Simulator"
            />
          }
          pageMap={await getPageMap()}
          docsRepositoryBase="https://github.com/Giuseppe-Compagnone/LWN-Simulator"
          editLink="Edit this page on GitHub"
          sidebar={{ defaultMenuCollapseLevel: 1, autoCollapse: true }}
          footer={footer}
        >
          {children}
        </Layout>
      </body>
    </html>
  );
}
