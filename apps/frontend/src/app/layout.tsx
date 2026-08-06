import { Footer, Navbar, Sidebar } from "@/components";
import "../styles/main.scss";
import "@lwn-simulator/ui-components/styles.css";
import "material-symbols/index.css";
import { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import cn from "classnames";

export const metadata: Metadata = {
  title: "LWN Simulator",
  icons: ["./icon.png"],
};

const inter = Inter({
  subsets: ["latin"],
  variable: "--primary-font",
});

const jetBrains = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--secondary-font",
});

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={cn("dark-theme", inter.variable, jetBrains.variable)}>
        <Navbar />
        <Sidebar>
          {children}
          <Footer />
        </Sidebar>
      </body>
    </html>
  );
}
