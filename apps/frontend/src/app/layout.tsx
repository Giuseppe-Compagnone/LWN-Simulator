import "../styles/main.scss";
import "@lwn-simulator/ui-components/styles.css";

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="dark-theme">{children}</body>
    </html>
  );
}
