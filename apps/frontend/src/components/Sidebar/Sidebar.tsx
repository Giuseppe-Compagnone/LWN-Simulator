import { SidebarProps } from "./Sidebar.types";

const Sidebar = (props: SidebarProps) => {
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
        <div className="sub-section active">
          <span className="material-symbols-outlined icon">dashboard</span>
          <span>Dashboard</span>
        </div>
        <div className="sub-section">
          <span className="material-symbols-outlined icon">dashboard</span>
          <span>Dashboard</span>
        </div>
        <div className="sub-section">
          <span className="material-symbols-outlined icon">dashboard</span>
          <span>Dashboard</span>
        </div>
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
      <div className="content">{props.children}</div>
    </div>
  );
};

export default Sidebar;
