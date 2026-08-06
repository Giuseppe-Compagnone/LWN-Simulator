"use client";

import { Logo, LogoSize } from "@lwn-simulator/ui-components";
import { NavbarProps } from "./Navbar.types";
import Link from "next/link";

const Navbar = (props: NavbarProps) => {
  return (
    <header className="navbar">
      <div className="app-name">
        <Logo size={LogoSize.Sm} />
        <strong>LWN Simulator</strong>
      </div>
      <div className="pages">
        <nav className="current">
          <Link href="/">lorem</Link>
        </nav>
        <nav>
          <Link href="/1">lorem</Link>
        </nav>
        <nav>
          <Link href="/2">lorem</Link>
        </nav>
        <nav>
          <Link href="/3">lorem</Link>
        </nav>
      </div>
    </header>
  );
};

export default Navbar;
