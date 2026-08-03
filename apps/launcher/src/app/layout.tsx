import type { Metadata } from "next";
import "@lwn-simulator/ui-components/styles.css";
import "./../styles/main.scss";

export const metadata: Metadata = {
  title: "LWN Simulator Desktop Launcher",
  icons: ["./icon.png"],
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="dark-theme">
        <main>{children}</main>
      </body>
    </html>
  );
}
