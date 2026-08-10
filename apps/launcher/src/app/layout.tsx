import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import "@lwn-simulator/ui-components/styles.css";
import "material-symbols/index.css";
import "./../styles/main.scss";
import { ToastContainer } from "react-toastify";
import cn from "classnames";
import { ThemeServiceProvider } from "@lwn-simulator/ui-components";
import { AppInfoServiceProvider } from "@lwn-simulator/sdk";

export const metadata: Metadata = {
  title: "LWN Simulator Desktop Launcher",
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
            <main>{children}</main>
            <ToastContainer />
          </AppInfoServiceProvider>
        </ThemeServiceProvider>
      </body>
    </html>
  );
}
