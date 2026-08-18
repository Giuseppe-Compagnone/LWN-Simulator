import "../styles/main.scss";
import "@lwn-simulator/ui-components/styles.css";
import "material-symbols/index.css";
import { Metadata } from "next";
import { ProvidersWrapper } from "@/components";
import { Inter, JetBrains_Mono } from "next/font/google";
import "react-toastify/ReactToastify.css";
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
        <ProvidersWrapper>{children}</ProvidersWrapper>
      </body>
    </html>
  );
}
