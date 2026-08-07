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
import cn from "classnames";
import { usePathname } from "next/navigation";

const Navbar = (props: NavbarProps) => {
  const pages = [
    { name: "Simulation", path: "/" },
    { name: "Hardware", path: "/hardware/" },
    { name: "Logs", path: "/logs/" },
  ];
  // Hooks
  const themeLogic = useThemeService();
  const pathname = usePathname();

  return (
    <header className="navbar">
      <div className="app-name">
        <Logo size={LogoSize.Sm} />
        <strong>LWN Simulator</strong>
      </div>
      <div className="pages">
        {pages.map((page, i) => {
          return (
            <nav key={i} className={cn(pathname === page.path && "current")}>
              <Link href={page.path}>{page.name}</Link>
            </nav>
          );
        })}
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
