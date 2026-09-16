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

  // Estraiamo la sezione principale del path in modo sicuro
  const mainSection = pathname?.split("/")[1] || "";
  const subSection = pathname?.split("/")[2] || "";

  // Recuperiamo le rotte corrispondenti o un array vuoto di fallback
  const currentRoutes = routes[mainSection] || [];

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

        {/* ✅ Ora .map() viene chiamato in sicurezza solo se le rotte esistono */}
        {currentRoutes.map((route, i) => {
          return (
            <Link
              className={cn(
                "sub-section",
                subSection === route.route && "active",
              )}
              key={i}
              href={`/${mainSection}/${route.route}`}
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
          rel="noreferrer"
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
