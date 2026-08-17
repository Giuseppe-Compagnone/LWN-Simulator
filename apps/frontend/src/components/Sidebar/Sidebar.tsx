"use client";

import { SidebarProps } from "./Sidebar.types";
import { usePathname } from "next/navigation";
import cn from "classnames";
import Link from "next/link";

const Sidebar = (props: SidebarProps) => {
  const routes: Record<string, Array<{ route: string; icon: string }>> = {
    simulation: [{ route: "dashboard", icon: "dashboard" }],
    hardware: [{ route: "devices", icon: "sensors" }],
    logs: [],
  };

  // Hooks
  const pathname = usePathname();

  return (
    <div className="sidebar">
      <aside>
        <div className="status">
          <span className="material-symbols-outlined icon">sensors</span>
          {true ? (
            <div className="active">Simulation active</div>
          ) : (
            <div className="inactive">Simulation inactive</div>
          )}
        </div>
        {pathname.split("/")[1] &&
          routes[pathname.split("/")[1]].map((route, i) => {
            return (
              <Link
                className={cn(
                  "sub-section",
                  (pathname.split("/")[2] || "") === route.route && "active",
                )}
                key={i}
                href={`/${pathname.split("/")[1]}/${route.route}`}
              >
                <span className="material-symbols-outlined icon">
                  {route.icon}
                </span>
                <span>{route.route}</span>
              </Link>
            );
          })}
        <hr />
        <a
          href="https://giuseppe-compagnone.github.io/LWN-Simulator/docs/"
          target="_blank"
          className="sub-section"
        >
          <span className="material-symbols-outlined icon">menu_book</span>
          <span>Documentation</span>
        </a>
      </aside>
      <main className="content">{props.children}</main>
    </div>
  );
};

export default Sidebar;
