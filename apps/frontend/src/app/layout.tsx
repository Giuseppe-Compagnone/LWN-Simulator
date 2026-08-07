import "../styles/main.scss";
import "@lwn-simulator/ui-components/styles.css";
import "material-symbols/index.css";
import { Metadata } from "next";
import { Footer, Navbar, Sidebar } from "@/components";
import { ThemeServiceProvider } from "@/services";

export const metadata: Metadata = {
  title: "LWN Simulator",
  icons: ["./icon.png"],
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <ThemeServiceProvider>
          <Navbar />
          <Sidebar>
            {children}
            <Footer />
          </Sidebar>
        </ThemeServiceProvider>
      </body>
    </html>
  );
}
