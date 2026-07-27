import { Footer, Layout, Navbar } from "nextra-theme-docs";
import { Head } from "nextra/components";
import { getPageMap } from "nextra/page-map";
import type { Metadata } from "next";
import type { ReactNode } from "react";
import "./globals.css";

export const metadata: Metadata = {
  title: "Nextra 4",
  description: "Nextra 4 is here.",
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
              logo={<h1 className="text-2xl">LWN Simulator</h1>}
              projectLink="https://github.com/Giuseppe-Compagnone/LWN-Simulator"
            />
          }
          pageMap={await getPageMap()}
          docsRepositoryBase="https://giuseppe-compagnone.github.io/LWN-Simulator/"
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
