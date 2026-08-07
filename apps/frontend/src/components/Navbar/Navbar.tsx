"use client";

import {
  Button,
  ButtonLayout,
  ButtonType,
  Logo,
  LogoSize,
} from "@lwn-simulator/ui-components";
import { NavbarProps } from "./Navbar.types";
import Link from "next/link";
import { Theme, useThemeService } from "@lwn-simulator/ui-components";

const Navbar = (props: NavbarProps) => {
  // Hooks
  const themeLogic = useThemeService();

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
      <Button
        value={
          <span className="material-symbols-outlined icon">
            {themeLogic.theme === Theme.Light ? "bedtime" : "sunny"}
          </span>
        }
        layout={ButtonLayout.Icon}
        type={ButtonType.Outlined}
        onClick={() => {
          themeLogic.setTheme(
            themeLogic.theme === Theme.Light ? Theme.Dark : Theme.Light,
          );
        }}
      />
    </header>
  );
};

export default Navbar;
