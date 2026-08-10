import "../styles/main.scss";
import "@lwn-simulator/ui-components/styles.css";
import "material-symbols/index.css";
import { Metadata } from "next";
import { Footer, Navbar, Sidebar } from "@/components";
import { Inter, JetBrains_Mono } from "next/font/google";
import { ThemeServiceProvider } from "@lwn-simulator/ui-components";
import { AppInfoServiceProvider } from "@lwn-simulator/sdk";
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
      <body className={cn(inter.variable, jetBrains.variable)}>
        <ThemeServiceProvider>
          <AppInfoServiceProvider
            baseUrl={`${process.env.NODE_ENV === "development" ? "http://localhost:8080" : window.location.origin}/api`}
          >
            <Navbar />
            <Sidebar>
              {children}
              <Footer />
            </Sidebar>
          </AppInfoServiceProvider>
        </ThemeServiceProvider>
      </body>
    </html>
  );
}
